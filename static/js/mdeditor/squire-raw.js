/* Copyright © 2011-2015 by Neil Jenkins. MIT Licensed. */

( function ( doc, undefined ) {

"use strict";

var DOCUMENT_POSITION_PRECEDING = 2; // Node.DOCUMENT_POSITION_PRECEDING
var ELEMENT_NODE = 1;                // Node.ELEMENT_NODE;
var TEXT_NODE = 3;                   // Node.TEXT_NODE;
var DOCUMENT_NODE = 9;               // Node.DOCUMENT_NODE;
var DOCUMENT_FRAGMENT_NODE = 11;     // Node.DOCUMENT_FRAGMENT_NODE;
var SHOW_ELEMENT = 1;                // NodeFilter.SHOW_ELEMENT;
var SHOW_TEXT = 4;                   // NodeFilter.SHOW_TEXT;

var START_TO_START = 0; // Range.START_TO_START
var START_TO_END = 1;   // Range.START_TO_END
var END_TO_END = 2;     // Range.END_TO_END
var END_TO_START = 3;   // Range.END_TO_START

var HIGHLIGHT_CLASS = 'highlight';
var COLOUR_CLASS = 'colour';
var FONT_FAMILY_CLASS = 'font';
var FONT_SIZE_CLASS = 'size';

var ZWS = '\u200B';

var win = doc.defaultView;

var ua = navigator.userAgent;

var isAndroid = /Android/.test( ua );
var isIOS = /iP(?:ad|hone|od)/.test( ua );
var isMac = /Mac OS X/.test( ua );
var isWin = /Windows NT/.test( ua );

var isGecko = /Gecko\//.test( ua );
var isIElt11 = /Trident\/[456]\./.test( ua );
var isPresto = !!win.opera;
var isEdge = /Edge\//.test( ua );
var isWebKit = !isEdge && /WebKit\//.test( ua );
var isIE = /Trident\/[4567]\./.test( ua );

var ctrlKey = isMac ? 'meta-' : 'ctrl-';

var useTextFixer = isIElt11 || isPresto;
var cantFocusEmptyTextNodes = isIElt11 || isWebKit;
var losesSelectionOnBlur = isIElt11;

var canObserveMutations = typeof MutationObserver !== 'undefined';
var canWeakMap = typeof WeakMap !== 'undefined';

// Use [^ \t\r\n] instead of \S so that nbsp does not count as white-space
var notWS = /[^ \t\r\n]/;

var indexOf = Array.prototype.indexOf;

// Polyfill for FF3.5
if ( !Object.create ) {
    Object.create = function ( proto ) {
        var F = function () {};
        F.prototype = proto;
        return new F();
    };
}

/*
    Native TreeWalker is buggy in IE and Opera:
    * IE9/10 sometimes throw errors when calling TreeWalker#nextNode or
      TreeWalker#previousNode. No way to feature detect this.
    * Some versions of Opera have a bug in TreeWalker#previousNode which makes
      it skip to the wrong node.

    Rather than risk further bugs, it's easiest just to implement our own
    (subset) of the spec in all browsers.
*/

var typeToBitArray = {
    // ELEMENT_NODE
    1: 1,
    // ATTRIBUTE_NODE
    2: 2,
    // TEXT_NODE
    3: 4,
    // COMMENT_NODE
    8: 128,
    // DOCUMENT_NODE
    9: 256,
    // DOCUMENT_FRAGMENT_NODE
    11: 1024
};

function TreeWalker ( root, nodeType, filter ) {
    this.root = this.currentNode = root;
    this.nodeType = nodeType;
    this.filter = filter;
}

TreeWalker.prototype.nextNode = function () {
    var current = this.currentNode,
        root = this.root,
        nodeType = this.nodeType,
        filter = this.filter,
        node;
    while ( true ) {
        node = current.firstChild;
        while ( !node && current ) {
            if ( current === root ) {
                break;
            }
            node = current.nextSibling;
            if ( !node ) { current = current.parentNode; }
        }
        if ( !node ) {
            return null;
        }
        if ( ( typeToBitArray[ node.nodeType ] & nodeType ) &&
                filter( node ) ) {
            this.currentNode = node;
            return node;
        }
        current = node;
    }
};

TreeWalker.prototype.previousNode = function () {
    var current = this.currentNode,
        root = this.root,
        nodeType = this.nodeType,
        filter = this.filter,
        node;
    while ( true ) {
        if ( current === root ) {
            return null;
        }
        node = current.previousSibling;
        if ( node ) {
            while ( current = node.lastChild ) {
                node = current;
            }
        } else {
            node = current.parentNode;
        }
        if ( !node ) {
            return null;
        }
        if ( ( typeToBitArray[ node.nodeType ] & nodeType ) &&
                filter( node ) ) {
            this.currentNode = node;
            return node;
        }
        current = node;
    }
};

// Previous node in post-order.
TreeWalker.prototype.previousPONode = function () {
    var current = this.currentNode,
        root = this.root,
        nodeType = this.nodeType,
        filter = this.filter,
        node;
    while ( true ) {
        node = current.lastChild;
        while ( !node && current ) {
            if ( current === root ) {
                break;
            }
            node = current.previousSibling;
            if ( !node ) { current = current.parentNode; }
        }
        if ( !node ) {
            return null;
        }
        if ( ( typeToBitArray[ node.nodeType ] & nodeType ) &&
                filter( node ) ) {
            this.currentNode = node;
            return node;
        }
        current = node;
    }
};

var inlineNodeNames  = /^(?:#text|A(?:BBR|CRONYM)?|B(?:R|D[IO])?|C(?:ITE|ODE)|D(?:ATA|EL|FN)|EM|FONT|HR|I(?:FRAME|MG|NPUT|NS)?|KBD|Q|R(?:P|T|UBY)|S(?:AMP|MALL|PAN|TR(?:IKE|ONG)|U[BP])?|TIME|U|VAR|WBR)$/;

var leafNodeNames = {
    BR: 1,
    HR: 1,
    IFRAME: 1,
    IMG: 1,
    INPUT: 1
};

function every ( nodeList, fn ) {
    var l = nodeList.length;
    while ( l-- ) {
        if ( !fn( nodeList[l] ) ) {
            return false;
        }
    }
    return true;
}

// ---

var UNKNOWN = 0;
var INLINE = 1;
var BLOCK = 2;
var CONTAINER = 3;

var nodeCategoryCache = canWeakMap ? new WeakMap() : null;

function isLeaf ( node ) {
    return node.nodeType === ELEMENT_NODE && !!leafNodeNames[ node.nodeName ];
}
function getNodeCategory ( node ) {
    switch ( node.nodeType ) {
    case TEXT_NODE:
        return INLINE;
    case ELEMENT_NODE:
    case DOCUMENT_FRAGMENT_NODE:
        if ( canWeakMap && nodeCategoryCache.has( node ) ) {
            return nodeCategoryCache.get( node );
        }
        break;
    default:
        return UNKNOWN;
    }

    var nodeCategory;
    if ( !every( node.childNodes, isInline ) ) {
        // Malformed HTML can have block tags inside inline tags. Need to treat
        // these as containers rather than inline. See #239.
        nodeCategory = CONTAINER;
    } else if ( inlineNodeNames.test( node.nodeName ) ) {
        nodeCategory = INLINE;
    } else {
        nodeCategory = BLOCK;
    }
    if ( canWeakMap ) {
        nodeCategoryCache.set( node, nodeCategory );
    }
    return nodeCategory;
}
function isInline ( node ) {
    return getNodeCategory( node ) === INLINE;
}
function isBlock ( node ) {
    return getNodeCategory( node ) === BLOCK;
}
function isContainer ( node ) {
    return getNodeCategory( node ) === CONTAINER;
}

function getBlockWalker ( node, root ) {
    var walker = new TreeWalker( root, SHOW_ELEMENT, isBlock );
    walker.currentNode = node;
    return walker;
}
function getPreviousBlock ( node, root ) {
    node = getBlockWalker( node, root ).previousNode();
    return node !== root ? node : null;
}
function getNextBlock ( node, root ) {
    node = getBlockWalker( node, root ).nextNode();
    return node !== root ? node : null;
}

function isEmptyBlock ( block ) {
    return !block.textContent && !block.querySelector( 'IMG' );
}

function areAlike ( node, node2 ) {
    return !isLeaf( node ) && (
        node.nodeType === node2.nodeType &&
        node.nodeName === node2.nodeName &&
        node.nodeName !== 'A' &&
        node.className === node2.className &&
        ( ( !node.style && !node2.style ) ||
          node.style.cssText === node2.style.cssText )
    );
}
function hasTagAttributes ( node, tag, attributes ) {
    if ( node.nodeName !== tag ) {
        return false;
    }
    for ( var attr in attributes ) {
        if ( node.getAttribute( attr ) !== attributes[ attr ] ) {
            return false;
        }
    }
    return true;
}
function getNearest ( node, root, tag, attributes ) {
    while ( node && node !== root ) {
        if ( hasTagAttributes( node, tag, attributes ) ) {
            return node;
        }
        node = node.parentNode;
    }
    return null;
}
function isOrContains ( parent, node ) {
    while ( node ) {
        if ( node === parent ) {
            return true;
        }
        node = node.parentNode;
    }
    return false;
}

function getPath ( node, root ) {
    var path = '';
    var id, className, classNames, dir;
    if ( node && node !== root ) {
        path = getPath( node.parentNode, root );
        if ( node.nodeType === ELEMENT_NODE ) {
            path += ( path ? '>' : '' ) + node.nodeName;
            if ( id = node.id ) {
                path += '#' + id;
            }
            if ( className = node.className.trim() ) {
                classNames = className.split( /\s\s*/ );
                classNames.sort();
                path += '.';
                path += classNames.join( '.' );
            }
            if ( dir = node.dir ) {
                path += '[dir=' + dir + ']';
            }
            if ( classNames ) {
                if ( indexOf.call( classNames, HIGHLIGHT_CLASS ) > -1 ) {
                    path += '[backgroundColor=' +
                        node.style.backgroundColor.replace( / /g,'' ) + ']';
                }
                if ( indexOf.call( classNames, COLOUR_CLASS ) > -1 ) {
                    path += '[color=' +
                        node.style.color.replace( / /g,'' ) + ']';
                }
                if ( indexOf.call( classNames, FONT_FAMILY_CLASS ) > -1 ) {
                    path += '[fontFamily=' +
                        node.style.fontFamily.replace( / /g,'' ) + ']';
                }
                if ( indexOf.call( classNames, FONT_SIZE_CLASS ) > -1 ) {
                    path += '[fontSize=' + node.style.fontSize + ']';
                }
            }
        }
    }
    return path;
}

function getLength ( node ) {
    var nodeType = node.nodeType;
    return nodeType === ELEMENT_NODE || nodeType === DOCUMENT_FRAGMENT_NODE ?
        node.childNodes.length : node.length || 0;
}

function detach ( node ) {
    var parent = node.parentNode;
    if ( parent ) {
        parent.removeChild( node );
    }
    return node;
}
function replaceWith ( node, node2 ) {
    var parent = node.parentNode;
    if ( parent ) {
        parent.replaceChild( node2, node );
    }
}
function empty ( node ) {
    var frag = node.ownerDocument.createDocumentFragment(),
        childNodes = node.childNodes,
        l = childNodes ? childNodes.length : 0;
    while ( l-- ) {
        frag.appendChild( node.firstChild );
    }
    return frag;
}

function createElement ( doc, tag, props, children ) {
    var el = doc.createElement( tag ),
        attr, value, i, l;
    if ( props instanceof Array ) {
        children = props;
        props = null;
    }
    if ( props ) {
        for ( attr in props ) {
            value = props[ attr ];
            if ( value !== undefined ) {
                el.setAttribute( attr, props[ attr ] );
            }
        }
    }
    if ( children ) {
        for ( i = 0, l = children.length; i < l; i += 1 ) {
            el.appendChild( children[i] );
        }
    }
    return el;
}

function fixCursor ( node, root ) {
    // In Webkit and Gecko, block level elements are collapsed and
    // unfocussable if they have no content. To remedy this, a <BR> must be
    // inserted. In Opera and IE, we just need a textnode in order for the
    // cursor to appear.
    var self = root.__squire__;
    var doc = node.ownerDocument;
    var originalNode = node;
    var fixer, child;

    if ( node === root ) {
        if ( !( child = node.firstChild ) || child.nodeName === 'BR' ) {
            fixer = self.createDefaultBlock();
            if ( child ) {
                node.replaceChild( fixer, child );
            }
            else {
                node.appendChild( fixer );
            }
            node = fixer;
            fixer = null;
        }
    }

    if ( node.nodeType === TEXT_NODE ) {
        return originalNode;
    }

    if ( isInline( node ) ) {
        child = node.firstChild;
        while ( cantFocusEmptyTextNodes && child &&
                child.nodeType === TEXT_NODE && !child.data ) {
            node.removeChild( child );
            child = node.firstChild;
        }
        if ( !child ) {
            if ( cantFocusEmptyTextNodes ) {
                fixer = doc.createTextNode( ZWS );
                self._didAddZWS();
            } else {
                fixer = doc.createTextNode( '' );
            }
        }
    } else {
        if ( useTextFixer ) {
            while ( node.nodeType !== TEXT_NODE && !isLeaf( node ) ) {
                child = node.firstChild;
                if ( !child ) {
                    fixer = doc.createTextNode( '' );
                    break;
                }
                node = child;
            }
            if ( node.nodeType === TEXT_NODE ) {
                // Opera will collapse the block element if it contains
                // just spaces (but not if it contains no data at all).
                if ( /^ +$/.test( node.data ) ) {
                    node.data = '';
                }
            } else if ( isLeaf( node ) ) {
                node.parentNode.insertBefore( doc.createTextNode( '' ), node );
            }
        }
        else if ( !node.querySelector( 'BR' ) ) {
            fixer = createElement( doc, 'BR' );
            while ( ( child = node.lastElementChild ) && !isInline( child ) ) {
                node = child;
            }
        }
    }
    if ( fixer ) {
        try {
            node.appendChild( fixer );
        } catch ( error ) {
            self.didError({
                name: 'Squire: fixCursor – ' + error,
                message: 'Parent: ' + node.nodeName + '/' + node.innerHTML +
                    ' appendChild: ' + fixer.nodeName
            });
        }
    }

    return originalNode;
}

// Recursively examine container nodes and wrap any inline children.
function fixContainer ( container, root ) {
    var children = container.childNodes;
    var doc = container.ownerDocument;
    var wrapper = null;
    var i, l, child, isBR;
    var config = root.__squire__._config;

    for ( i = 0, l = children.length; i < l; i += 1 ) {
        child = children[i];
        isBR = child.nodeName === 'BR';
        if ( !isBR && isInline( child ) ) {
            if ( !wrapper ) {
                 wrapper = createElement( doc,
                    config.blockTag, config.blockAttributes );
            }
            wrapper.appendChild( child );
            i -= 1;
            l -= 1;
        } else if ( isBR || wrapper ) {
            if ( !wrapper ) {
                wrapper = createElement( doc,
                    config.blockTag, config.blockAttributes );
            }
            fixCursor( wrapper, root );
            if ( isBR ) {
                container.replaceChild( wrapper, child );
            } else {
                container.insertBefore( wrapper, child );
                i += 1;
                l += 1;
            }
            wrapper = null;
        }
        if ( isContainer( child ) ) {
            fixContainer( child, root );
        }
    }
    if ( wrapper ) {
        container.appendChild( fixCursor( wrapper, root ) );
    }
    return container;
}

function split ( node, offset, stopNode, root ) {
    var nodeType = node.nodeType,
        parent, clone, next;
    if ( nodeType === TEXT_NODE && node !== stopNode ) {
        return split(
            node.parentNode, node.splitText( offset ), stopNode, root );
    }
    if ( nodeType === ELEMENT_NODE ) {
        if ( typeof( offset ) === 'number' ) {
            offset = offset < node.childNodes.length ?
                node.childNodes[ offset ] : null;
        }
        if ( node === stopNode ) {
            return offset;
        }

        // Clone node without children
        parent = node.parentNode;
        clone = node.cloneNode( false );

        // Add right-hand siblings to the clone
        while ( offset ) {
            next = offset.nextSibling;
            clone.appendChild( offset );
            offset = next;
        }

        // Maintain li numbering if inside a quote.
        if ( node.nodeName === 'OL' &&
                getNearest( node, root, 'BLOCKQUOTE' ) ) {
            clone.start = ( +node.start || 1 ) + node.childNodes.length - 1;
        }

        // DO NOT NORMALISE. This may undo the fixCursor() call
        // of a node lower down the tree!

        // We need something in the element in order for the cursor to appear.
        fixCursor( node, root );
        fixCursor( clone, root );

        // Inject clone after original node
        if ( next = node.nextSibling ) {
            parent.insertBefore( clone, next );
        } else {
            parent.appendChild( clone );
        }

        // Keep on splitting up the tree
        return split( parent, clone, stopNode, root );
    }
    return offset;
}

function _mergeInlines ( node, fakeRange ) {
    var children = node.childNodes,
        l = children.length,
        frags = [],
        child, prev, len;
    while ( l-- ) {
        child = children[l];
        prev = l && children[ l - 1 ];
        if ( l && isInline( child ) && areAlike( child, prev ) &&
                !leafNodeNames[ child.nodeName ] ) {
            if ( fakeRange.startContainer === child ) {
                fakeRange.startContainer = prev;
                fakeRange.startOffset += getLength( prev );
            }
            if ( fakeRange.endContainer === child ) {
                fakeRange.endContainer = prev;
                fakeRange.endOffset += getLength( prev );
            }
            if ( fakeRange.startContainer === node ) {
                if ( fakeRange.startOffset > l ) {
                    fakeRange.startOffset -= 1;
                }
                else if ( fakeRange.startOffset === l ) {
                    fakeRange.startContainer = prev;
                    fakeRange.startOffset = getLength( prev );
                }
            }
            if ( fakeRange.endContainer === node ) {
                if ( fakeRange.endOffset > l ) {
                    fakeRange.endOffset -= 1;
                }
                else if ( fakeRange.endOffset === l ) {
                    fakeRange.endContainer = prev;
                    fakeRange.endOffset = getLength( prev );
                }
            }
            detach( child );
            if ( child.nodeType === TEXT_NODE ) {
                prev.appendData( child.data );
            }
            else {
                frags.push( empty( child ) );
            }
        }
        else if ( child.nodeType === ELEMENT_NODE ) {
            len = frags.length;
            while ( len-- ) {
                child.appendChild( frags.pop() );
            }
            _mergeInlines( child, fakeRange );
        }
    }
}

function mergeInlines ( node, range ) {
    if ( node.nodeType === TEXT_NODE ) {
        node = node.parentNode;
    }
    if ( node.nodeType === ELEMENT_NODE ) {
        var fakeRange = {
            startContainer: range.startContainer,
            startOffset: range.startOffset,
            endContainer: range.endContainer,
            endOffset: range.endOffset
        };
        _mergeInlines( node, fakeRange );
        range.setStart( fakeRange.startContainer, fakeRange.startOffset );
        range.setEnd( fakeRange.endContainer, fakeRange.endOffset );
    }
}

function mergeWithBlock ( block, next, range, root ) {
    var container = next;
    var parent, last, offset;
    while ( ( parent = container.parentNode ) &&
            parent !== root &&
            parent.nodeType === ELEMENT_NODE &&
            parent.childNodes.length === 1 ) {
        container = parent;
    }
    detach( container );

    offset = block.childNodes.length;

    // Remove extra <BR> fixer if present.
    last = block.lastChild;
    if ( last && last.nodeName === 'BR' ) {
        block.removeChild( last );
        offset -= 1;
    }

    block.appendChild( empty( next ) );

    range.setStart( block, offset );
    range.collapse( true );
    mergeInlines( block, range );

    // Opera inserts a BR if you delete the last piece of text
    // in a block-level element. Unfortunately, it then gets
    // confused when setting the selection subsequently and
    // refuses to accept the range that finishes just before the
    // BR. Removing the BR fixes the bug.
    // Steps to reproduce bug: Type "a-b-c" (where - is return)
    // then backspace twice. The cursor goes to the top instead
    // of after "b".
    if ( isPresto && ( last = block.lastChild ) && last.nodeName === 'BR' ) {
        block.removeChild( last );
    }
}

function mergeContainers ( node, root ) {
    var prev = node.previousSibling,
        first = node.firstChild,
        doc = node.ownerDocument,
        isListItem = ( node.nodeName === 'LI' ),
        needsFix, block;

    // Do not merge LIs, unless it only contains a UL
    if ( isListItem && ( !first || !/^[OU]L$/.test( first.nodeName ) ) ) {
        return;
    }

    if ( prev && areAlike( prev, node ) ) {
        if ( !isContainer( prev ) ) {
            if ( isListItem ) {
                block = createElement( doc, 'DIV' );
                block.appendChild( empty( prev ) );
                prev.appendChild( block );
            } else {
                return;
            }
        }
        detach( node );
        needsFix = !isContainer( node );
        prev.appendChild( empty( node ) );
        if ( needsFix ) {
            fixContainer( prev, root );
        }
        if ( first ) {
            mergeContainers( first, root );
        }
    } else if ( isListItem ) {
        prev = createElement( doc, 'DIV' );
        node.insertBefore( prev, first );
        fixCursor( prev, root );
    }
}

var getNodeBefore = function ( node, offset ) {
    var children = node.childNodes;
    while ( offset && node.nodeType === ELEMENT_NODE ) {
        node = children[ offset - 1 ];
        children = node.childNodes;
        offset = children.length;
    }
    return node;
};

var getNodeAfter = function ( node, offset ) {
    if ( node.nodeType === ELEMENT_NODE ) {
        var children = node.childNodes;
        if ( offset < children.length ) {
            node = children[ offset ];
        } else {
            while ( node && !node.nextSibling ) {
                node = node.parentNode;
            }
            if ( node ) { node = node.nextSibling; }
        }
    }
    return node;
};

// ---

var insertNodeInRange = function ( range, node ) {
    // Insert at start.
    var startContainer = range.startContainer,
        startOffset = range.startOffset,
        endContainer = range.endContainer,
        endOffset = range.endOffset,
        parent, children, childCount, afterSplit;

    // If part way through a text node, split it.
    if ( startContainer.nodeType === TEXT_NODE ) {
        parent = startContainer.parentNode;
        children = parent.childNodes;
        if ( startOffset === startContainer.length ) {
            startOffset = indexOf.call( children, startContainer ) + 1;
            if ( range.collapsed ) {
                endContainer = parent;
                endOffset = startOffset;
            }
        } else {
            if ( startOffset ) {
                afterSplit = startContainer.splitText( startOffset );
                if ( endContainer === startContainer ) {
                    endOffset -= startOffset;
                    endContainer = afterSplit;
                }
                else if ( endContainer === parent ) {
                    endOffset += 1;
                }
                startContainer = afterSplit;
            }
            startOffset = indexOf.call( children, startContainer );
        }
        startContainer = parent;
    } else {
        children = startContainer.childNodes;
    }

    childCount = children.length;

    if ( startOffset === childCount ) {
        startContainer.appendChild( node );
    } else {
        startContainer.insertBefore( node, children[ startOffset ] );
    }

    if ( startContainer === endContainer ) {
        endOffset += children.length - childCount;
    }

    range.setStart( startContainer, startOffset );
    range.setEnd( endContainer, endOffset );
};

var extractContentsOfRange = function ( range, common, root ) {
    var startContainer = range.startContainer,
        startOffset = range.startOffset,
        endContainer = range.endContainer,
        endOffset = range.endOffset;

    if ( !common ) {
        common = range.commonAncestorContainer;
    }

    if ( common.nodeType === TEXT_NODE ) {
        common = common.parentNode;
    }

    var endNode = split( endContainer, endOffset, common, root ),
        startNode = split( startContainer, startOffset, common, root ),
        frag = common.ownerDocument.createDocumentFragment(),
        next, before, after;

    // End node will be null if at end of child nodes list.
    while ( startNode !== endNode ) {
        next = startNode.nextSibling;
        frag.appendChild( startNode );
        startNode = next;
    }

    startContainer = common;
    startOffset = endNode ?
        indexOf.call( common.childNodes, endNode ) :
        common.childNodes.length;

    // Merge text nodes if adjacent. IE10 in particular will not focus
    // between two text nodes
    after = common.childNodes[ startOffset ];
    before = after && after.previousSibling;
    if ( before &&
            before.nodeType === TEXT_NODE &&
            after.nodeType === TEXT_NODE ) {
        startContainer = before;
        startOffset = before.length;
        before.appendData( after.data );
        detach( after );
    }

    range.setStart( startContainer, startOffset );
    range.collapse( true );

    fixCursor( common, root );

    return frag;
};

var deleteContentsOfRange = function ( range, root ) {
    var startBlock = getStartBlockOfRange( range, root );
    var endBlock = getEndBlockOfRange( range, root );
    var needsMerge = ( startBlock !== endBlock );
    var frag, child;

    // Move boundaries up as much as possible without exiting block,
    // to reduce need to split.
    moveRangeBoundariesDownTree( range );
    moveRangeBoundariesUpTree( range, startBlock, endBlock, root );

    // Remove selected range
    frag = extractContentsOfRange( range, null, root );

    // Move boundaries back down tree as far as possible.
    moveRangeBoundariesDownTree( range );

    // If we split into two different blocks, merge the blocks.
    if ( needsMerge ) {
        // endBlock will have been split, so need to refetch
        endBlock = getEndBlockOfRange( range, root );
        if ( startBlock && endBlock && startBlock !== endBlock ) {
            mergeWithBlock( startBlock, endBlock, range, root );
        }
    }

    // Ensure block has necessary children
    if ( startBlock ) {
        fixCursor( startBlock, root );
    }

    // Ensure root has a block-level element in it.
    child = root.firstChild;
    if ( !child || child.nodeName === 'BR' ) {
        fixCursor( root, root );
        range.selectNodeContents( root.firstChild );
    } else {
        range.collapse( true );
    }
    return frag;
};

// ---

// Contents of range will be deleted.
// After method, range will be around inserted content
var insertTreeFragmentIntoRange = function ( range, frag, root ) {
    var node, block, blockContentsAfterSplit, stopPoint, container, offset;
    var firstBlockInFrag, nodeAfterSplit, nodeBeforeSplit, tempRange;

    // Fixup content: ensure no top-level inline, and add cursor fix elements.
    fixContainer( frag, root );
    node = frag;
    while ( ( node = getNextBlock( node, root ) ) ) {
        fixCursor( node, root );
    }

    // Delete any selected content.
    if ( !range.collapsed ) {
        deleteContentsOfRange( range, root );
    }

    // Move range down into text nodes.
    moveRangeBoundariesDownTree( range );
    range.collapse( false ); // collapse to end

    // Where will we split up to? First blockquote parent, otherwise root.
    stopPoint = getNearest( range.endContainer, root, 'BLOCKQUOTE' ) || root;

    // Merge the contents of the first block in the frag with the focused block.
    // If there are contents in the block after the focus point, collect this
    // up to insert in the last block later
    block = getStartBlockOfRange( range, root );
    firstBlockInFrag = getNextBlock( frag, frag );
    if ( block && firstBlockInFrag &&
            // Don't merge table cells or PRE elements into block
            !getNearest( firstBlockInFrag, frag, 'PRE' ) &&
            !getNearest( firstBlockInFrag, frag, 'TABLE' ) ) {
        moveRangeBoundariesUpTree( range, block, block, root );
        range.collapse( true ); // collapse to start
        container = range.endContainer;
        offset = range.endOffset;
        // Remove trailing <br> – we don't want this considered content to be
        // inserted again later
        cleanupBRs( block, root, false );
        if ( isInline( container ) ) {
            // Split up to block parent.
            nodeAfterSplit = split(
                container, offset, getPreviousBlock( container, root ), root );
            container = nodeAfterSplit.parentNode;
            offset = indexOf.call( container.childNodes, nodeAfterSplit );
        }
        if ( /*isBlock( container ) && */offset !== getLength( container ) ) {
            // Collect any inline contents of the block after the range point
            blockContentsAfterSplit =
                root.ownerDocument.createDocumentFragment();
            while ( ( node = container.childNodes[ offset ] ) ) {
                blockContentsAfterSplit.appendChild( node );
            }
        }
        // And merge the first block in.
        mergeWithBlock( container, firstBlockInFrag, range, root );

        // And where we will insert
        offset = indexOf.call( container.parentNode.childNodes, container ) + 1;
        container = container.parentNode;
        range.setEnd( container, offset );
    }

    // Is there still any content in the fragment?
    if ( getLength( frag ) ) {
        moveRangeBoundariesUpTree( range, stopPoint, stopPoint, root );
        // Now split after block up to blockquote (if a parent) or root
        nodeAfterSplit = split(
            range.endContainer, range.endOffset, stopPoint, root );
        nodeBeforeSplit = nodeAfterSplit ?
            nodeAfterSplit.previousSibling :
            stopPoint.lastChild;
        stopPoint.insertBefore( frag, nodeAfterSplit );
        if ( nodeAfterSplit ) {
            range.setEndBefore( nodeAfterSplit );
        } else {
            range.setEnd( stopPoint, getLength( stopPoint ) );
        }
        block = getEndBlockOfRange( range, root );

        // Get a reference that won't be invalidated if we merge containers.
        moveRangeBoundariesDownTree( range );
        container = range.endContainer;
        offset = range.endOffset;

        // Merge inserted containers with edges of split
        if ( nodeAfterSplit && isContainer( nodeAfterSplit ) ) {
            mergeContainers( nodeAfterSplit, root );
        }
        nodeAfterSplit = nodeBeforeSplit && nodeBeforeSplit.nextSibling;
        if ( nodeAfterSplit && isContainer( nodeAfterSplit ) ) {
            mergeContainers( nodeAfterSplit, root );
        }
        range.setEnd( container, offset );
    }

    // Insert inline content saved from before.
    if ( blockContentsAfterSplit ) {
        tempRange = range.cloneRange();
        mergeWithBlock( block, blockContentsAfterSplit, tempRange, root );
        range.setEnd( tempRange.endContainer, tempRange.endOffset );
    }
    moveRangeBoundariesDownTree( range );
};

// ---

var isNodeContainedInRange = function ( range, node, partial ) {
    var nodeRange = node.ownerDocument.createRange();

    nodeRange.selectNode( node );

    if ( partial ) {
        // Node must not finish before range starts or start after range
        // finishes.
        var nodeEndBeforeStart = ( range.compareBoundaryPoints(
                END_TO_START, nodeRange ) > -1 ),
            nodeStartAfterEnd = ( range.compareBoundaryPoints(
                START_TO_END, nodeRange ) < 1 );
        return ( !nodeEndBeforeStart && !nodeStartAfterEnd );
    }
    else {
        // Node must start after range starts and finish before range
        // finishes
        var nodeStartAfterStart = ( range.compareBoundaryPoints(
                START_TO_START, nodeRange ) < 1 ),
            nodeEndBeforeEnd = ( range.compareBoundaryPoints(
                END_TO_END, nodeRange ) > -1 );
        return ( nodeStartAfterStart && nodeEndBeforeEnd );
    }
};

var moveRangeBoundariesDownTree = function ( range ) {
    var startContainer = range.startContainer,
        startOffset = range.startOffset,
        endContainer = range.endContainer,
        endOffset = range.endOffset,
        maySkipBR = true,
        child;

    while ( startContainer.nodeType !== TEXT_NODE ) {
        child = startContainer.childNodes[ startOffset ];
        if ( !child || isLeaf( child ) ) {
            break;
        }
        startContainer = child;
        startOffset = 0;
    }
    if ( endOffset ) {
        while ( endContainer.nodeType !== TEXT_NODE ) {
            child = endContainer.childNodes[ endOffset - 1 ];
            if ( !child || isLeaf( child ) ) {
                if ( maySkipBR && child && child.nodeName === 'BR' ) {
                    endOffset -= 1;
                    maySkipBR = false;
                    continue;
                }
                break;
            }
            endContainer = child;
            endOffset = getLength( endContainer );
        }
    } else {
        while ( endContainer.nodeType !== TEXT_NODE ) {
            child = endContainer.firstChild;
            if ( !child || isLeaf( child ) ) {
                break;
            }
            endContainer = child;
        }
    }

    // If collapsed, this algorithm finds the nearest text node positions
    // *outside* the range rather than inside, but also it flips which is
    // assigned to which.
    if ( range.collapsed ) {
        range.setStart( endContainer, endOffset );
        range.setEnd( startContainer, startOffset );
    } else {
        range.setStart( startContainer, startOffset );
        range.setEnd( endContainer, endOffset );
    }
};

var moveRangeBoundariesUpTree = function ( range, startMax, endMax, root ) {
    var startContainer = range.startContainer;
    var startOffset = range.startOffset;
    var endContainer = range.endContainer;
    var endOffset = range.endOffset;
    var maySkipBR = true;
    var parent;

    if ( !startMax ) {
        startMax = range.commonAncestorContainer;
    }
    if ( !endMax ) {
        endMax = startMax;
    }

    while ( !startOffset &&
            startContainer !== startMax &&
            startContainer !== root ) {
        parent = startContainer.parentNode;
        startOffset = indexOf.call( parent.childNodes, startContainer );
        startContainer = parent;
    }

    while ( true ) {
        if ( maySkipBR &&
                endContainer.nodeType !== TEXT_NODE &&
                endContainer.childNodes[ endOffset ] &&
                endContainer.childNodes[ endOffset ].nodeName === 'BR' ) {
            endOffset += 1;
            maySkipBR = false;
        }
        if ( endContainer === endMax ||
                endContainer === root ||
                endOffset !== getLength( endContainer ) ) {
            break;
        }
        parent = endContainer.parentNode;
        endOffset = indexOf.call( parent.childNodes, endContainer ) + 1;
        endContainer = parent;
    }

    range.setStart( startContainer, startOffset );
    range.setEnd( endContainer, endOffset );
};

// Returns the first block at least partially contained by the range,
// or null if no block is contained by the range.
var getStartBlockOfRange = function ( range, root ) {
    var container = range.startContainer,
        block;

    // If inline, get the containing block.
    if ( isInline( container ) ) {
        block = getPreviousBlock( container, root );
    } else if ( container !== root && isBlock( container ) ) {
        block = container;
    } else {
        block = getNodeBefore( container, range.startOffset );
        block = getNextBlock( block, root );
    }
    // Check the block actually intersects the range
    return block && isNodeContainedInRange( range, block, true ) ? block : null;
};

// Returns the last block at least partially contained by the range,
// or null if no block is contained by the range.
var getEndBlockOfRange = function ( range, root ) {
    var container = range.endContainer,
        block, child;

    // If inline, get the containing block.
    if ( isInline( container ) ) {
        block = getPreviousBlock( container, root );
    } else if ( container !== root && isBlock( container ) ) {
        block = container;
    } else {
        block = getNodeAfter( container, range.endOffset );
        if ( !block || !isOrContains( root, block ) ) {
            block = root;
            while ( child = block.lastChild ) {
                block = child;
            }
        }
        block = getPreviousBlock( block, root );
    }
    // Check the block actually intersects the range
    return block && isNodeContainedInRange( range, block, true ) ? block : null;
};

var contentWalker = new TreeWalker( null,
    SHOW_TEXT|SHOW_ELEMENT,
    function ( node ) {
        return node.nodeType === TEXT_NODE ?
            notWS.test( node.data ) :
            node.nodeName === 'IMG';
    }
);

var rangeDoesStartAtBlockBoundary = function ( range, root ) {
    var startContainer = range.startContainer;
    var startOffset = range.startOffset;
    var nodeAfterCursor;

    // If in the middle or end of a text node, we're not at the boundary.
    contentWalker.root = null;
    if ( startContainer.nodeType === TEXT_NODE ) {
        if ( startOffset ) {
            return false;
        }
        nodeAfterCursor = startContainer;
    } else {
        nodeAfterCursor = getNodeAfter( startContainer, startOffset );
        if ( nodeAfterCursor && !isOrContains( root, nodeAfterCursor ) ) {
            nodeAfterCursor = null;
        }
        // The cursor was right at the end of the document
        if ( !nodeAfterCursor ) {
            nodeAfterCursor = getNodeBefore( startContainer, startOffset );
            if ( nodeAfterCursor.nodeType === TEXT_NODE &&
                    nodeAfterCursor.length ) {
                return false;
            }
        }
    }

    // Otherwise, look for any previous content in the same block.
    contentWalker.currentNode = nodeAfterCursor;
    contentWalker.root = getStartBlockOfRange( range, root );

    return !contentWalker.previousNode();
};

var rangeDoesEndAtBlockBoundary = function ( range, root ) {
    var endContainer = range.endContainer,
        endOffset = range.endOffset,
        length;

    // If in a text node with content, and not at the end, we're not
    // at the boundary
    contentWalker.root = null;
    if ( endContainer.nodeType === TEXT_NODE ) {
        length = endContainer.data.length;
        if ( length && endOffset < length ) {
            return false;
        }
        contentWalker.currentNode = endContainer;
    } else {
        contentWalker.currentNode = getNodeBefore( endContainer, endOffset );
    }

    // Otherwise, look for any further content in the same block.
    contentWalker.root = getEndBlockOfRange( range, root );

    return !contentWalker.nextNode();
};

var expandRangeToBlockBoundaries = function ( range, root ) {
    var start = getStartBlockOfRange( range, root ),
        end = getEndBlockOfRange( range, root ),
        parent;

    if ( start && end ) {
        parent = start.parentNode;
        range.setStart( parent, indexOf.call( parent.childNodes, start ) );
        parent = end.parentNode;
        range.setEnd( parent, indexOf.call( parent.childNodes, end ) + 1 );
    }
};

var keys = {
    8: 'backspace',
    9: 'tab',
    13: 'enter',
    32: 'space',
    33: 'pageup',
    34: 'pagedown',
    37: 'left',
    39: 'right',
    46: 'delete',
    219: '[',
    221: ']'
};

// Ref: http://unixpapa.com/js/key.html
var onKey = function ( event ) {
    var code = event.keyCode,
        key = keys[ code ],
        modifiers = '',
        range = this.getSelection();

    if ( event.defaultPrevented ) {
        return;
    }

    if ( !key ) {
        key = String.fromCharCode( code ).toLowerCase();
        // Only reliable for letters and numbers
        if ( !/^[A-Za-z0-9]$/.test( key ) ) {
            key = '';
        }
    }

    // On keypress, delete and '.' both have event.keyCode 46
    // Must check event.which to differentiate.
    if ( isPresto && event.which === 46 ) {
        key = '.';
    }

    // Function keys
    if ( 111 < code && code < 124 ) {
        key = 'f' + ( code - 111 );
    }

    // We need to apply the backspace/delete handlers regardless of
    // control key modifiers.
    if ( key !== 'backspace' && key !== 'delete' ) {
        if ( event.altKey  ) { modifiers += 'alt-'; }
        if ( event.ctrlKey ) { modifiers += 'ctrl-'; }
        if ( event.metaKey ) { modifiers += 'meta-'; }
    }
    // However, on Windows, shift-delete is apparently "cut" (WTF right?), so
    // we want to let the browser handle shift-delete.
    if ( event.shiftKey ) { modifiers += 'shift-'; }

    key = modifiers + key;

    if ( this._keyHandlers[ key ] ) {
        this._keyHandlers[ key ]( this, event, range );
    } else if ( key.length === 1 && !range.collapsed ) {
        // Record undo checkpoint.
        this.saveUndoState( range );
        // Delete the selection
        deleteContentsOfRange( range, this._root );
        this._ensureBottomLine();
        this.setSelection( range );
        this._updatePath( range, true );
    }
};

var mapKeyTo = function ( method ) {
    return function ( self, event ) {
        event.preventDefault();
        self[ method ]();
    };
};

var mapKeyToFormat = function ( tag, remove ) {
    remove = remove || null;
    return function ( self, event ) {
        event.preventDefault();
        var range = self.getSelection();
        if ( self.hasFormat( tag, null, range ) ) {
            self.changeFormat( null, { tag: tag }, range );
        } else {
            self.changeFormat( { tag: tag }, remove, range );
        }
    };
};

// If you delete the content inside a span with a font styling, Webkit will
// replace it with a <font> tag (!). If you delete all the text inside a
// link in Opera, it won't delete the link. Let's make things consistent. If
// you delete all text inside an inline tag, remove the inline tag.
var afterDelete = function ( self, range ) {
    try {
        if ( !range ) { range = self.getSelection(); }
        var node = range.startContainer,
            parent;
        // Climb the tree from the focus point while we are inside an empty
        // inline element
        if ( node.nodeType === TEXT_NODE ) {
            node = node.parentNode;
        }
        parent = node;
        while ( isInline( parent ) &&
                ( !parent.textContent || parent.textContent === ZWS ) ) {
            node = parent;
            parent = node.parentNode;
        }
        // If focused in empty inline element
        if ( node !== parent ) {
            // Move focus to just before empty inline(s)
            range.setStart( parent,
                indexOf.call( parent.childNodes, node ) );
            range.collapse( true );
            // Remove empty inline(s)
            parent.removeChild( node );
            // Fix cursor in block
            if ( !isBlock( parent ) ) {
                parent = getPreviousBlock( parent, self._root );
            }
            fixCursor( parent, self._root );
            // Move cursor into text node
            moveRangeBoundariesDownTree( range );
        }
        // If you delete the last character in the sole <div> in Chrome,
        // it removes the div and replaces it with just a <br> inside the
        // root. Detach the <br>; the _ensureBottomLine call will insert a new
        // block.
        if ( node === self._root &&
                ( node = node.firstChild ) && node.nodeName === 'BR' ) {
            detach( node );
        }
        self._ensureBottomLine();
        self.setSelection( range );
        self._updatePath( range, true );
    } catch ( error ) {
        self.didError( error );
    }
};

var keyHandlers = {
    enter: function ( self, event, range ) {
        var root = self._root;
        var block, parent, nodeAfterSplit;

        // We handle this ourselves
        event.preventDefault();

        // Save undo checkpoint and add any links in the preceding section.
        // Remove any zws so we don't think there's content in an empty
        // block.
        self._recordUndoState( range );
        addLinks( range.startContainer, root, self );
        self._removeZWS();
        self._getRangeAndRemoveBookmark( range );

        // Selected text is overwritten, therefore delete the contents
        // to collapse selection.
        if ( !range.collapsed ) {
            deleteContentsOfRange( range, root );
        }

        block = getStartBlockOfRange( range, root );

        // If this is a malformed bit of document or in a table;
        // just play it safe and insert a <br>.
        if ( !block || /^T[HD]$/.test( block.nodeName ) ) {
            // If inside an <a>, move focus out
            parent = getNearest( range.endContainer, root, 'A' );
            if ( parent ) {
                parent = parent.parentNode;
                moveRangeBoundariesUpTree( range, parent, parent, root );
                range.collapse( false );
            }
            insertNodeInRange( range, self.createElement( 'BR' ) );
            range.collapse( false );
            self.setSelection( range );
            self._updatePath( range, true );
            return;
        }

        // If in a list, we'll split the LI instead.
        if ( parent = getNearest( block, root, 'LI' ) ) {
            block = parent;
        }

        if ( isEmptyBlock( block ) ) {
            // Break list
            if ( getNearest( block, root, 'UL' ) ||
                    getNearest( block, root, 'OL' ) ) {
                return self.decreaseListLevel( range );
            }
            // Break blockquote
            else if ( getNearest( block, root, 'BLOCKQUOTE' ) ) {
                return self.modifyBlocks( removeBlockQuote, range );
            }
        }

        // Otherwise, split at cursor point.
        nodeAfterSplit = splitBlock( self, block,
            range.startContainer, range.startOffset );

        // Clean up any empty inlines if we hit enter at the beginning of the
        // block
        removeZWS( block );
        removeEmptyInlines( block );
        fixCursor( block, root );

        // Focus cursor
        // If there's a <b>/<i> etc. at the beginning of the split
        // make sure we focus inside it.
        while ( nodeAfterSplit.nodeType === ELEMENT_NODE ) {
            var child = nodeAfterSplit.firstChild,
                next;

            // Don't continue links over a block break; unlikely to be the
            // desired outcome.
            if ( nodeAfterSplit.nodeName === 'A' &&
                    ( !nodeAfterSplit.textContent ||
                        nodeAfterSplit.textContent === ZWS ) ) {
                child = self._doc.createTextNode( '' );
                replaceWith( nodeAfterSplit, child );
                nodeAfterSplit = child;
                break;
            }

            while ( child && child.nodeType === TEXT_NODE && !child.data ) {
                next = child.nextSibling;
                if ( !next || next.nodeName === 'BR' ) {
                    break;
                }
                detach( child );
                child = next;
            }

            // 'BR's essentially don't count; they're a browser hack.
            // If you try to select the contents of a 'BR', FF will not let
            // you type anything!
            if ( !child || child.nodeName === 'BR' ||
                    ( child.nodeType === TEXT_NODE && !isPresto ) ) {
                break;
            }
            nodeAfterSplit = child;
        }
        range = self._createRange( nodeAfterSplit, 0 );
        self.setSelection( range );
        self._updatePath( range, true );
    },
    backspace: function ( self, event, range ) {
        var root = self._root;
        self._removeZWS();
        // Record undo checkpoint.
        self.saveUndoState( range );
        // If not collapsed, delete contents
        if ( !range.collapsed ) {
            event.preventDefault();
            deleteContentsOfRange( range, root );
            afterDelete( self, range );
        }
        // If at beginning of block, merge with previous
        else if ( rangeDoesStartAtBlockBoundary( range, root ) ) {
            event.preventDefault();
            var current = getStartBlockOfRange( range, root );
            var previous;
            if ( !current ) {
                return;
            }
            // In case inline data has somehow got between blocks.
            fixContainer( current.parentNode, root );
            // Now get previous block
            previous = getPreviousBlock( current, root );
            // Must not be at the very beginning of the text area.
            if ( previous ) {
                // If not editable, just delete whole block.
                if ( !previous.isContentEditable ) {
                    detach( previous );
                    return;
                }
                // Otherwise merge.
                mergeWithBlock( previous, current, range, root );
                // If deleted line between containers, merge newly adjacent
                // containers.
                current = previous.parentNode;
                while ( current !== root && !current.nextSibling ) {
                    current = current.parentNode;
                }
                if ( current !== root && ( current = current.nextSibling ) ) {
                    mergeContainers( current, root );
                }
                self.setSelection( range );
            }
            // If at very beginning of text area, allow backspace
            // to break lists/blockquote.
            else if ( current ) {
                // Break list
                if ( getNearest( current, root, 'UL' ) ||
                        getNearest( current, root, 'OL' ) ) {
                    return self.decreaseListLevel( range );
                }
                // Break blockquote
                else if ( getNearest( current, root, 'BLOCKQUOTE' ) ) {
                    return self.modifyBlocks( decreaseBlockQuoteLevel, range );
                }
                self.setSelection( range );
                self._updatePath( range, true );
            }
        }
        // Otherwise, leave to browser but check afterwards whether it has
        // left behind an empty inline tag.
        else {
            self.setSelection( range );
            setTimeout( function () { afterDelete( self ); }, 0 );
        }
    },
    'delete': function ( self, event, range ) {
        var root = self._root;
        var current, next, originalRange,
            cursorContainer, cursorOffset, nodeAfterCursor;
        self._removeZWS();
        // Record undo checkpoint.
        self.saveUndoState( range );
        // If not collapsed, delete contents
        if ( !range.collapsed ) {
            event.preventDefault();
            deleteContentsOfRange( range, root );
            afterDelete( self, range );
        }
        // If at end of block, merge next into this block
        else if ( rangeDoesEndAtBlockBoundary( range, root ) ) {
            event.preventDefault();
            current = getStartBlockOfRange( range, root );
            if ( !current ) {
                return;
            }
            // In case inline data has somehow got between blocks.
            fixContainer( current.parentNode, root );
            // Now get next block
            next = getNextBlock( current, root );
            // Must not be at the very end of the text area.
            if ( next ) {
                // If not editable, just delete whole block.
                if ( !next.isContentEditable ) {
                    detach( next );
                    return;
                }
                // Otherwise merge.
                mergeWithBlock( current, next, range, root );
                // If deleted line between containers, merge newly adjacent
                // containers.
                next = current.parentNode;
                while ( next !== root && !next.nextSibling ) {
                    next = next.parentNode;
                }
                if ( next !== root && ( next = next.nextSibling ) ) {
                    mergeContainers( next, root );
                }
                self.setSelection( range );
                self._updatePath( range, true );
            }
        }
        // Otherwise, leave to browser but check afterwards whether it has
        // left behind an empty inline tag.
        else {
            // But first check if the cursor is just before an IMG tag. If so,
            // delete it ourselves, because the browser won't if it is not
            // inline.
            originalRange = range.cloneRange();
            moveRangeBoundariesUpTree( range, root, root, root );
            cursorContainer = range.endContainer;
            cursorOffset = range.endOffset;
            if ( cursorContainer.nodeType === ELEMENT_NODE ) {
                nodeAfterCursor = cursorContainer.childNodes[ cursorOffset ];
                if ( nodeAfterCursor && nodeAfterCursor.nodeName === 'IMG' ) {
                    event.preventDefault();
                    detach( nodeAfterCursor );
                    moveRangeBoundariesDownTree( range );
                    afterDelete( self, range );
                    return;
                }
            }
            self.setSelection( originalRange );
            setTimeout( function () { afterDelete( self ); }, 0 );
        }
    },
    tab: function ( self, event, range ) {
        var root = self._root;
        var node, parent;
        self._removeZWS();
        // If no selection and at start of block
        if ( range.collapsed && rangeDoesStartAtBlockBoundary( range, root ) ) {
            node = getStartBlockOfRange( range, root );
            // Iterate through the block's parents
            while ( ( parent = node.parentNode ) ) {
                // If we find a UL or OL (so are in a list, node must be an LI)
                if ( parent.nodeName === 'UL' || parent.nodeName === 'OL' ) {
                    // Then increase the list level
                    event.preventDefault();
                    self.increaseListLevel( range );
                    break;
                }
                node = parent;
            }
        }
    },
    'shift-tab': function ( self, event, range ) {
        var root = self._root;
        var node;
        self._removeZWS();
        // If no selection and at start of block
        if ( range.collapsed && rangeDoesStartAtBlockBoundary( range, root ) ) {
            // Break list
            node = range.startContainer;
            if ( getNearest( node, root, 'UL' ) ||
                    getNearest( node, root, 'OL' ) ) {
                event.preventDefault();
                self.decreaseListLevel( range );
            }
        }
    },
    space: function ( self, _, range ) {
        var node, parent;
        self._recordUndoState( range );
        addLinks( range.startContainer, self._root, self );
        self._getRangeAndRemoveBookmark( range );

        // If the cursor is at the end of a link (<a>foo|</a>) then move it
        // outside of the link (<a>foo</a>|) so that the space is not part of
        // the link text.
        node = range.endContainer;
        parent = node.parentNode;
        if ( range.collapsed && parent.nodeName === 'A' &&
                !node.nextSibling && range.endOffset === getLength( node ) ) {
            range.setStartAfter( parent );
        }
        // Delete the selection if not collapsed
        else if ( !range.collapsed ) {
            deleteContentsOfRange( range, self._root );
            self._ensureBottomLine();
            self.setSelection( range );
            self._updatePath( range, true );
        }

        self.setSelection( range );
    },
    left: function ( self ) {
        self._removeZWS();
    },
    right: function ( self ) {
        self._removeZWS();
    }
};

// Firefox pre v29 incorrectly handles Cmd-left/Cmd-right on Mac:
// it goes back/forward in history! Override to do the right
// thing.
// https://bugzilla.mozilla.org/show_bug.cgi?id=289384
if ( isMac && isGecko ) {
    keyHandlers[ 'meta-left' ] = function ( self, event ) {
        event.preventDefault();
        var sel = getWindowSelection( self );
        if ( sel && sel.modify ) {
            sel.modify( 'move', 'backward', 'lineboundary' );
        }
    };
    keyHandlers[ 'meta-right' ] = function ( self, event ) {
        event.preventDefault();
        var sel = getWindowSelection( self );
        if ( sel && sel.modify ) {
            sel.modify( 'move', 'forward', 'lineboundary' );
        }
    };
}

// System standard for page up/down on Mac is to just scroll, not move the
// cursor. On Linux/Windows, it should move the cursor, but some browsers don't
// implement this natively. Override to support it.
if ( !isMac ) {
    keyHandlers.pageup = function ( self ) {
        self.moveCursorToStart();
    };
    keyHandlers.pagedown = function ( self ) {
        self.moveCursorToEnd();
    };
}

keyHandlers[ ctrlKey + 'b' ] = mapKeyToFormat( 'B' );
keyHandlers[ ctrlKey + 'i' ] = mapKeyToFormat( 'I' );
keyHandlers[ ctrlKey + 'u' ] = mapKeyToFormat( 'U' );
keyHandlers[ ctrlKey + 'shift-7' ] = mapKeyToFormat( 'S' );
keyHandlers[ ctrlKey + 'shift-5' ] = mapKeyToFormat( 'SUB', { tag: 'SUP' } );
keyHandlers[ ctrlKey + 'shift-6' ] = mapKeyToFormat( 'SUP', { tag: 'SUB' } );
keyHandlers[ ctrlKey + 'shift-8' ] = mapKeyTo( 'makeUnorderedList' );
keyHandlers[ ctrlKey + 'shift-9' ] = mapKeyTo( 'makeOrderedList' );
keyHandlers[ ctrlKey + '[' ] = mapKeyTo( 'decreaseQuoteLevel' );
keyHandlers[ ctrlKey + ']' ] = mapKeyTo( 'increaseQuoteLevel' );
keyHandlers[ ctrlKey + 'y' ] = mapKeyTo( 'redo' );
keyHandlers[ ctrlKey + 'z' ] = mapKeyTo( 'undo' );
keyHandlers[ ctrlKey + 'shift-z' ] = mapKeyTo( 'redo' );

var fontSizes = {
    1: 10,
    2: 13,
    3: 16,
    4: 18,
    5: 24,
    6: 32,
    7: 48
};

var styleToSemantic = {
    backgroundColor: {
        regexp: notWS,
        replace: function ( doc, colour ) {
            return createElement( doc, 'SPAN', {
                'class': HIGHLIGHT_CLASS,
                style: 'background-color:' + colour
            });
        }
    },
    color: {
        regexp: notWS,
        replace: function ( doc, colour ) {
            return createElement( doc, 'SPAN', {
                'class': COLOUR_CLASS,
                style: 'color:' + colour
            });
        }
    },
    fontWeight: {
        regexp: /^bold|^700/i,
        replace: function ( doc ) {
            return createElement( doc, 'B' );
        }
    },
    fontStyle: {
        regexp: /^italic/i,
        replace: function ( doc ) {
            return createElement( doc, 'I' );
        }
    },
    fontFamily: {
        regexp: notWS,
        replace: function ( doc, family ) {
            return createElement( doc, 'SPAN', {
                'class': FONT_FAMILY_CLASS,
                style: 'font-family:' + family
            });
        }
    },
    fontSize: {
        regexp: notWS,
        replace: function ( doc, size ) {
            return createElement( doc, 'SPAN', {
                'class': FONT_SIZE_CLASS,
                style: 'font-size:' + size
            });
        }
    },
    textDecoration: {
        regexp: /^underline/i,
        replace: function ( doc ) {
            return createElement( doc, 'U' );
        }
    }
};

var replaceWithTag = function ( tag ) {
    return function ( node, parent ) {
        var el = createElement( node.ownerDocument, tag );
        parent.replaceChild( el, node );
        el.appendChild( empty( node ) );
        return el;
    };
};

var replaceStyles = function ( node, parent ) {
    var style = node.style;
    var doc = node.ownerDocument;
    var attr, converter, css, newTreeBottom, newTreeTop, el;

    for ( attr in styleToSemantic ) {
        converter = styleToSemantic[ attr ];
        css = style[ attr ];
        if ( css && converter.regexp.test( css ) ) {
            el = converter.replace( doc, css );
            if ( !newTreeTop ) {
                newTreeTop = el;
            }
            if ( newTreeBottom ) {
                newTreeBottom.appendChild( el );
            }
            newTreeBottom = el;
            node.style[ attr ] = '';
        }
    }

    if ( newTreeTop ) {
        newTreeBottom.appendChild( empty( node ) );
        if ( node.nodeName === 'SPAN' ) {
            parent.replaceChild( newTreeTop, node );
        } else {
            node.appendChild( newTreeTop );
        }
    }

    return newTreeBottom || node;
};

var stylesRewriters = {
    P: replaceStyles,
    SPAN: replaceStyles,
    STRONG: replaceWithTag( 'B' ),
    EM: replaceWithTag( 'I' ),
    INS: replaceWithTag( 'U' ),
    STRIKE: replaceWithTag( 'S' ),
    FONT: function ( node, parent ) {
        var face = node.face,
            size = node.size,
            colour = node.color,
            doc = node.ownerDocument,
            fontSpan, sizeSpan, colourSpan,
            newTreeBottom, newTreeTop;
        if ( face ) {
            fontSpan = createElement( doc, 'SPAN', {
                'class': FONT_FAMILY_CLASS,
                style: 'font-family:' + face
            });
            newTreeTop = fontSpan;
            newTreeBottom = fontSpan;
        }
        if ( size ) {
            sizeSpan = createElement( doc, 'SPAN', {
                'class': FONT_SIZE_CLASS,
                style: 'font-size:' + fontSizes[ size ] + 'px'
            });
            if ( !newTreeTop ) {
                newTreeTop = sizeSpan;
            }
            if ( newTreeBottom ) {
                newTreeBottom.appendChild( sizeSpan );
            }
            newTreeBottom = sizeSpan;
        }
        if ( colour && /^#?([\dA-F]{3}){1,2}$/i.test( colour ) ) {
            if ( colour.charAt( 0 ) !== '#' ) {
                colour = '#' + colour;
            }
            colourSpan = createElement( doc, 'SPAN', {
                'class': COLOUR_CLASS,
                style: 'color:' + colour
            });
            if ( !newTreeTop ) {
                newTreeTop = colourSpan;
            }
            if ( newTreeBottom ) {
                newTreeBottom.appendChild( colourSpan );
            }
            newTreeBottom = colourSpan;
        }
        if ( !newTreeTop ) {
            newTreeTop = newTreeBottom = createElement( doc, 'SPAN' );
        }
        parent.replaceChild( newTreeTop, node );
        newTreeBottom.appendChild( empty( node ) );
        return newTreeBottom;
    },
    TT: function ( node, parent ) {
        var el = createElement( node.ownerDocument, 'SPAN', {
            'class': FONT_FAMILY_CLASS,
            style: 'font-family:menlo,consolas,"courier new",monospace'
        });
        parent.replaceChild( el, node );
        el.appendChild( empty( node ) );
        return el;
    }
};

var allowedBlock = /^(?:A(?:DDRESS|RTICLE|SIDE|UDIO)|BLOCKQUOTE|CAPTION|D(?:[DLT]|IV)|F(?:IGURE|IGCAPTION|OOTER)|H[1-6]|HEADER|L(?:ABEL|EGEND|I)|O(?:L|UTPUT)|P(?:RE)?|SECTION|T(?:ABLE|BODY|D|FOOT|H|HEAD|R)|COL(?:GROUP)?|UL)$/;

var blacklist = /^(?:HEAD|META|STYLE)/;

var walker = new TreeWalker( null, SHOW_TEXT|SHOW_ELEMENT, function () {
    return true;
});

/*
    Two purposes:

    1. Remove nodes we don't want, such as weird <o:p> tags, comment nodes
       and whitespace nodes.
    2. Convert inline tags into our preferred format.
*/
var cleanTree = function cleanTree ( node, preserveWS ) {
    var children = node.childNodes,
        nonInlineParent, i, l, child, nodeName, nodeType, rewriter, childLength,
        startsWithWS, endsWithWS, data, sibling;

    nonInlineParent = node;
    while ( isInline( nonInlineParent ) ) {
        nonInlineParent = nonInlineParent.parentNode;
    }
    walker.root = nonInlineParent;

    for ( i = 0, l = children.length; i < l; i += 1 ) {
        child = children[i];
        nodeName = child.nodeName;
        nodeType = child.nodeType;
        rewriter = stylesRewriters[ nodeName ];
        if ( nodeType === ELEMENT_NODE ) {
            childLength = child.childNodes.length;
            if ( rewriter ) {
                child = rewriter( child, node );
            } else if ( blacklist.test( nodeName ) ) {
                node.removeChild( child );
                i -= 1;
                l -= 1;
                continue;
            } else if ( !allowedBlock.test( nodeName ) && !isInline( child ) ) {
                i -= 1;
                l += childLength - 1;
                node.replaceChild( empty( child ), child );
                continue;
            }
            if ( childLength ) {
                cleanTree( child, preserveWS || ( nodeName === 'PRE' ) );
            }
        } else {
            if ( nodeType === TEXT_NODE ) {
                data = child.data;
                startsWithWS = !notWS.test( data.charAt( 0 ) );
                endsWithWS = !notWS.test( data.charAt( data.length - 1 ) );
                if ( preserveWS || ( !startsWithWS && !endsWithWS ) ) {
                    continue;
                }
                // Iterate through the nodes; if we hit some other content
                // before the start of a new block we don't trim
                if ( startsWithWS ) {
                    walker.currentNode = child;
                    while ( sibling = walker.previousPONode() ) {
                        nodeName = sibling.nodeName;
                        if ( nodeName === 'IMG' ||
                                ( nodeName === '#text' &&
                                    notWS.test( sibling.data ) ) ) {
                            break;
                        }
                        if ( !isInline( sibling ) ) {
                            sibling = null;
                            break;
                        }
                    }
                    data = data.replace( /^[ \t\r\n]+/g, sibling ? ' ' : '' );
                }
                if ( endsWithWS ) {
                    walker.currentNode = child;
                    while ( sibling = walker.nextNode() ) {
                        if ( nodeName === 'IMG' ||
                                ( nodeName === '#text' &&
                                    notWS.test( sibling.data ) ) ) {
                            break;
                        }
                        if ( !isInline( sibling ) ) {
                            sibling = null;
                            break;
                        }
                    }
                    data = data.replace( /[ \t\r\n]+$/g, sibling ? ' ' : '' );
                }
                if ( data ) {
                    child.data = data;
                    continue;
                }
            }
            node.removeChild( child );
            i -= 1;
            l -= 1;
        }
    }
    return node;
};

// ---

var removeEmptyInlines = function removeEmptyInlines ( node ) {
    var children = node.childNodes,
        l = children.length,
        child;
    while ( l-- ) {
        child = children[l];
        if ( child.nodeType === ELEMENT_NODE && !isLeaf( child ) ) {
            removeEmptyInlines( child );
            if ( isInline( child ) && !child.firstChild ) {
                node.removeChild( child );
            }
        } else if ( child.nodeType === TEXT_NODE && !child.data ) {
            node.removeChild( child );
        }
    }
};

// ---

var notWSTextNode = function ( node ) {
    return node.nodeType === ELEMENT_NODE ?
        node.nodeName === 'BR' :
        notWS.test( node.data );
};
var isLineBreak = function ( br, isLBIfEmptyBlock ) {
    var block = br.parentNode;
    var walker;
    while ( isInline( block ) ) {
        block = block.parentNode;
    }
    walker = new TreeWalker(
        block, SHOW_ELEMENT|SHOW_TEXT, notWSTextNode );
    walker.currentNode = br;
    return !!walker.nextNode() ||
        ( isLBIfEmptyBlock && !walker.previousNode() );
};

// <br> elements are treated specially, and differently depending on the
// browser, when in rich text editor mode. When adding HTML from external
// sources, we must remove them, replacing the ones that actually affect
// line breaks by wrapping the inline text in a <div>. Browsers that want <br>
// elements at the end of each block will then have them added back in a later
// fixCursor method call.
var cleanupBRs = function ( node, root, keepForBlankLine ) {
    var brs = node.querySelectorAll( 'BR' );
    var brBreaksLine = [];
    var l = brs.length;
    var i, br, parent;

    // Must calculate whether the <br> breaks a line first, because if we
    // have two <br>s next to each other, after the first one is converted
    // to a block split, the second will be at the end of a block and
    // therefore seem to not be a line break. But in its original context it
    // was, so we should also convert it to a block split.
    for ( i = 0; i < l; i += 1 ) {
        brBreaksLine[i] = isLineBreak( brs[i], keepForBlankLine );
    }
    while ( l-- ) {
        br = brs[l];
        // Cleanup may have removed it
        parent = br.parentNode;
        if ( !parent ) { continue; }
        // If it doesn't break a line, just remove it; it's not doing
        // anything useful. We'll add it back later if required by the
        // browser. If it breaks a line, wrap the content in div tags
        // and replace the brs.
        if ( !brBreaksLine[l] ) {
            detach( br );
        } else if ( !isInline( parent ) ) {
            fixContainer( parent, root );
        }
    }
};

// The (non-standard but supported enough) innerText property is based on the
// render tree in Firefox and possibly other browsers, so we must insert the
// DOM node into the document to ensure the text part is correct.
var setClipboardData = function ( clipboardData, node, root ) {
    var body = node.ownerDocument.body;
    var html, text;

    // Firefox will add an extra new line for BRs at the end of block when
    // calculating innerText, even though they don't actually affect display.
    // So we need to remove them first.
    cleanupBRs( node, root, true );

    node.setAttribute( 'style',
        'position:fixed;overflow:hidden;bottom:100%;right:100%;' );
    body.appendChild( node );
    html = node.innerHTML;
    text = node.innerText || node.textContent;

    // Firefox (and others?) returns unix line endings (\n) even on Windows.
    // If on Windows, normalise to \r\n, since Notepad and some other crappy
    // apps do not understand just \n.
    if ( isWin ) {
        text = text.replace( /\r?\n/g, '\r\n' );
    }

    clipboardData.setData( 'text/html', html );
    clipboardData.setData( 'text/plain', text );

    body.removeChild( node );
};

var onCut = function ( event ) {
    var clipboardData = event.clipboardData;
    var range = this.getSelection();
    var root = this._root;
    var self = this;
    var startBlock, endBlock, copyRoot, contents, parent, newContents, node;

    // Nothing to do
    if ( range.collapsed ) {
        event.preventDefault();
        return;
    }

    // Save undo checkpoint
    this.saveUndoState( range );

    // Edge only seems to support setting plain text as of 2016-03-11.
    // Mobile Safari flat out doesn't work:
    // https://bugs.webkit.org/show_bug.cgi?id=143776
    if ( !isEdge && !isIOS && clipboardData ) {
        // Clipboard content should include all parents within block, or all
        // parents up to root if selection across blocks
        startBlock = getStartBlockOfRange( range, root );
        endBlock = getEndBlockOfRange( range, root );
        copyRoot = ( ( startBlock === endBlock ) && startBlock ) || root;
        // Extract the contents
        contents = deleteContentsOfRange( range, root );
        // Add any other parents not in extracted content, up to copy root
        parent = range.commonAncestorContainer;
        if ( parent.nodeType === TEXT_NODE ) {
            parent = parent.parentNode;
        }
        while ( parent && parent !== copyRoot ) {
            newContents = parent.cloneNode( false );
            newContents.appendChild( contents );
            contents = newContents;
            parent = parent.parentNode;
        }
        // Set clipboard data
        node = this.createElement( 'div' );
        node.appendChild( contents );
        setClipboardData( clipboardData, node, root );
        event.preventDefault();
    } else {
        setTimeout( function () {
            try {
                // If all content removed, ensure div at start of root.
                self._ensureBottomLine();
            } catch ( error ) {
                self.didError( error );
            }
        }, 0 );
    }

    this.setSelection( range );
};

var onCopy = function ( event ) {
    var clipboardData = event.clipboardData;
    var range = this.getSelection();
    var root = this._root;
    var startBlock, endBlock, copyRoot, contents, parent, newContents, node;

    // Edge only seems to support setting plain text as of 2016-03-11.
    // Mobile Safari flat out doesn't work:
    // https://bugs.webkit.org/show_bug.cgi?id=143776
    if ( !isEdge && !isIOS && clipboardData ) {
        // Clipboard content should include all parents within block, or all
        // parents up to root if selection across blocks
        startBlock = getStartBlockOfRange( range, root );
        endBlock = getEndBlockOfRange( range, root );
        copyRoot = ( ( startBlock === endBlock ) && startBlock ) || root;
        // Clone range to mutate, then move up as high as possible without
        // passing the copy root node.
        range = range.cloneRange();
        moveRangeBoundariesDownTree( range );
        moveRangeBoundariesUpTree( range, copyRoot, copyRoot, root );
        // Extract the contents
        contents = range.cloneContents();
        // Add any other parents not in extracted content, up to copy root
        parent = range.commonAncestorContainer;
        if ( parent.nodeType === TEXT_NODE ) {
            parent = parent.parentNode;
        }
        while ( parent && parent !== copyRoot ) {
            newContents = parent.cloneNode( false );
            newContents.appendChild( contents );
            contents = newContents;
            parent = parent.parentNode;
        }
        // Set clipboard data
        node = this.createElement( 'div' );
        node.appendChild( contents );
        setClipboardData( clipboardData, node, root );
        event.preventDefault();
    }
};

// Need to monitor for shift key like this, as event.shiftKey is not available
// in paste event.
function monitorShiftKey ( event ) {
    this.isShiftDown = event.shiftKey;
}

var onPaste = function ( event ) {
    var clipboardData = event.clipboardData;
    var items = clipboardData && clipboardData.items;
    var choosePlain = this.isShiftDown;
    var fireDrop = false;
    var hasImage = false;
    var plainItem = null;
    var self = this;
    var l, item, type, types, data;

    // Current HTML5 Clipboard interface
    // ---------------------------------
    // https://html.spec.whatwg.org/multipage/interaction.html

    // Edge only provides access to plain text as of 2016-03-11.
    if ( !isEdge && items ) {
        event.preventDefault();
        l = items.length;
        while ( l-- ) {
            item = items[l];
            type = item.type;
            if ( !choosePlain && type === 'text/html' ) {
                /*jshint loopfunc: true */
                item.getAsString( function ( html ) {
                    self.insertHTML( html, true );
                });
                /*jshint loopfunc: false */
                return;
            }
            if ( type === 'text/plain' ) {
                plainItem = item;
            }
            if ( !choosePlain && /^image\/.*/.test( type ) ) {
                hasImage = true;
            }
        }
        // Treat image paste as a drop of an image file.
        if ( hasImage ) {
            this.fireEvent( 'dragover', {
                dataTransfer: clipboardData,
                /*jshint loopfunc: true */
                preventDefault: function () {
                    fireDrop = true;
                }
                /*jshint loopfunc: false */
            });
            if ( fireDrop ) {
                this.fireEvent( 'drop', {
                    dataTransfer: clipboardData
                });
            }
        } else if ( plainItem ) {
            plainItem.getAsString( function ( text ) {
                self.insertPlainText( text, true );
            });
        }
        return;
    }

    // Old interface
    // -------------

    // Safari (and indeed many other OS X apps) copies stuff as text/rtf
    // rather than text/html; even from a webpage in Safari. The only way
    // to get an HTML version is to fallback to letting the browser insert
    // the content. Same for getting image data. *Sigh*.
    //
    // Firefox is even worse: it doesn't even let you know that there might be
    // an RTF version on the clipboard, but it will also convert to HTML if you
    // let the browser insert the content. I've filed
    // https://bugzilla.mozilla.org/show_bug.cgi?id=1254028
    types = clipboardData && clipboardData.types;
    if ( !isEdge && types && (
            indexOf.call( types, 'text/html' ) > -1 || (
                !isGecko &&
                indexOf.call( types, 'text/plain' ) > -1 &&
                indexOf.call( types, 'text/rtf' ) < 0 )
            )) {
        event.preventDefault();
        // Abiword on Linux copies a plain text and html version, but the HTML
        // version is the empty string! So always try to get HTML, but if none,
        // insert plain text instead. On iOS, Facebook (and possibly other
        // apps?) copy links as type text/uri-list, but also insert a **blank**
        // text/plain item onto the clipboard. Why? Who knows.
        if ( !choosePlain && ( data = clipboardData.getData( 'text/html' ) ) ) {
            this.insertHTML( data, true );
        } else if (
                ( data = clipboardData.getData( 'text/plain' ) ) ||
                ( data = clipboardData.getData( 'text/uri-list' ) ) ) {
            this.insertPlainText( data, true );
        }
        return;
    }

    // No interface. Includes all versions of IE :(
    // --------------------------------------------

    this._awaitingPaste = true;

    var body = this._doc.body,
        range = this.getSelection(),
        startContainer = range.startContainer,
        startOffset = range.startOffset,
        endContainer = range.endContainer,
        endOffset = range.endOffset;

    // We need to position the pasteArea in the visible portion of the screen
    // to stop the browser auto-scrolling.
    var pasteArea = this.createElement( 'DIV', {
        contenteditable: 'true',
        style: 'position:fixed; overflow:hidden; top:0; right:100%; width:1px; height:1px;'
    });
    body.appendChild( pasteArea );
    range.selectNodeContents( pasteArea );
    this.setSelection( range );

    // A setTimeout of 0 means this is added to the back of the
    // single javascript thread, so it will be executed after the
    // paste event.
    setTimeout( function () {
        try {
            // IE sometimes fires the beforepaste event twice; make sure it is
            // not run again before our after paste function is called.
            self._awaitingPaste = false;

            // Get the pasted content and clean
            var html = '',
                next = pasteArea,
                first, range;

            // #88: Chrome can apparently split the paste area if certain
            // content is inserted; gather them all up.
            while ( pasteArea = next ) {
                next = pasteArea.nextSibling;
                detach( pasteArea );
                // Safari and IE like putting extra divs around things.
                first = pasteArea.firstChild;
                if ( first && first === pasteArea.lastChild &&
                        first.nodeName === 'DIV' ) {
                    pasteArea = first;
                }
                html += pasteArea.innerHTML;
            }

            range = self._createRange(
                startContainer, startOffset, endContainer, endOffset );
            self.setSelection( range );

            if ( html ) {
                self.insertHTML( html, true );
            }
        } catch ( error ) {
            self.didError( error );
        }
    }, 0 );
};

// On Windows you can drag an drop text. We can't handle this ourselves, because
// as far as I can see, there's no way to get the drop insertion point. So just
// save an undo state and hope for the best.
var onDrop = function ( event ) {
    var types = event.dataTransfer.types;
    var l = types.length;
    var hasPlain = false;
    var hasHTML = false;
    while ( l-- ) {
        switch ( types[l] ) {
        case 'text/plain':
            hasPlain = true;
            break;
        case 'text/html':
            hasHTML = true;
            break;
        default:
            return;
        }
    }
    if ( hasHTML || hasPlain ) {
        this.saveUndoState();
    }
};

function mergeObjects ( base, extras, mayOverride ) {
    var prop, value;
    if ( !base ) {
        base = {};
    }
    if ( extras ) {
        for ( prop in extras ) {
            if ( mayOverride || !( prop in base ) ) {
                value = extras[ prop ];
                base[ prop ] = ( value && value.constructor === Object ) ?
                    mergeObjects( base[ prop ], value, mayOverride ) :
                    value;
            }
        }
    }
    return base;
}

function Squire ( root, config ) {
    if ( root.nodeType === DOCUMENT_NODE ) {
        root = root.body;
    }
    var doc = root.ownerDocument;
    var win = doc.defaultView;
    var mutation;

    this._win = win;
    this._doc = doc;
    this._root = root;

    this._events = {};

    this._isFocused = false;
    this._lastSelection = null;

    // IE loses selection state of iframe on blur, so make sure we
    // cache it just before it loses focus.
    if ( losesSelectionOnBlur ) {
        this.addEventListener( 'beforedeactivate', this.getSelection );
    }

    this._hasZWS = false;

    this._lastAnchorNode = null;
    this._lastFocusNode = null;
    this._path = '';
    this._willUpdatePath = false;

    if ( 'onselectionchange' in doc ) {
        this.addEventListener( 'selectionchange', this._updatePathOnEvent );
    } else {
        this.addEventListener( 'keyup', this._updatePathOnEvent );
        this.addEventListener( 'mouseup', this._updatePathOnEvent );
    }

    this._undoIndex = -1;
    this._undoStack = [];
    this._undoStackLength = 0;
    this._isInUndoState = false;
    this._ignoreChange = false;
    this._ignoreAllChanges = false;

    if ( canObserveMutations ) {
        mutation = new MutationObserver( this._docWasChanged.bind( this ) );
        mutation.observe( root, {
            childList: true,
            attributes: true,
            characterData: true,
            subtree: true
        });
        this._mutation = mutation;
    } else {
        this.addEventListener( 'keyup', this._keyUpDetectChange );
    }

    // On blur, restore focus except if the user taps or clicks to focus a
    // specific point. Can't actually use click event because focus happens
    // before click, so use mousedown/touchstart
    this._restoreSelection = false;
    this.addEventListener( 'blur', enableRestoreSelection );
    this.addEventListener( 'mousedown', disableRestoreSelection );
    this.addEventListener( 'touchstart', disableRestoreSelection );
    this.addEventListener( 'focus', restoreSelection );

    // IE sometimes fires the beforepaste event twice; make sure it is not run
    // again before our after paste function is called.
    this._awaitingPaste = false;
    this.addEventListener( isIElt11 ? 'beforecut' : 'cut', onCut );
    this.addEventListener( 'copy', onCopy );
    this.addEventListener( 'keydown', monitorShiftKey );
    this.addEventListener( 'keyup', monitorShiftKey );
    this.addEventListener( isIElt11 ? 'beforepaste' : 'paste', onPaste );
    this.addEventListener( 'drop', onDrop );

    // Opera does not fire keydown repeatedly.
    this.addEventListener( isPresto ? 'keypress' : 'keydown', onKey );

    // Add key handlers
    this._keyHandlers = Object.create( keyHandlers );

    // Override default properties
    this.setConfig( config );

    // Fix IE<10's buggy implementation of Text#splitText.
    // If the split is at the end of the node, it doesn't insert the newly split
    // node into the document, and sets its value to undefined rather than ''.
    // And even if the split is not at the end, the original node is removed
    // from the document and replaced by another, rather than just having its
    // data shortened.
    // We used to feature test for this, but then found the feature test would
    // sometimes pass, but later on the buggy behaviour would still appear.
    // I think IE10 does not have the same bug, but it doesn't hurt to replace
    // its native fn too and then we don't need yet another UA category.
    if ( isIElt11 ) {
        win.Text.prototype.splitText = function ( offset ) {
            var afterSplit = this.ownerDocument.createTextNode(
                    this.data.slice( offset ) ),
                next = this.nextSibling,
                parent = this.parentNode,
                toDelete = this.length - offset;
            if ( next ) {
                parent.insertBefore( afterSplit, next );
            } else {
                parent.appendChild( afterSplit );
            }
            if ( toDelete ) {
                this.deleteData( offset, toDelete );
            }
            return afterSplit;
        };
    }

    root.setAttribute( 'contenteditable', 'true' );

    // Remove Firefox's built-in controls
    try {
        doc.execCommand( 'enableObjectResizing', false, 'false' );
        doc.execCommand( 'enableInlineTableEditing', false, 'false' );
    } catch ( error ) {}

    root.__squire__ = this;

    // Need to register instance before calling setHTML, so that the fixCursor
    // function can lookup any default block tag options set.
    this.setHTML( '' );
}

var proto = Squire.prototype;

var sanitizeToDOMFragment = function ( html, isPaste, self ) {
    var doc = self._doc;
    var frag = html ? DOMPurify.sanitize( html, {
        ALLOW_UNKNOWN_PROTOCOLS: true,
        WHOLE_DOCUMENT: false,
        RETURN_DOM: true,
        RETURN_DOM_FRAGMENT: true
    }) : null;
    return frag ? doc.importNode( frag, true ) : doc.createDocumentFragment();
};

proto.setConfig = function ( config ) {
    config = mergeObjects({
        blockTag: 'DIV',
        blockAttributes: null,
        tagAttributes: {
            blockquote: null,
            ul: null,
            ol: null,
            li: null,
            a: null
        },
        leafNodeNames: leafNodeNames,
        undo: {
            documentSizeThreshold: -1, // -1 means no threshold
            undoLimit: -1 // -1 means no limit
        },
        isInsertedHTMLSanitized: true,
        isSetHTMLSanitized: true,
        sanitizeToDOMFragment:
            typeof DOMPurify !== 'undefined' && DOMPurify.isSupported ?
            sanitizeToDOMFragment : null

    }, config, true );

    // Users may specify block tag in lower case
    config.blockTag = config.blockTag.toUpperCase();

    this._config = config;

    return this;
};

proto.createElement = function ( tag, props, children ) {
    return createElement( this._doc, tag, props, children );
};

proto.createDefaultBlock = function ( children ) {
    var config = this._config;
    return fixCursor(
        this.createElement( config.blockTag, config.blockAttributes, children ),
        this._root
    );
};

proto.didError = function ( error ) {
    console.log( error );
};

proto.getDocument = function () {
    return this._doc;
};
proto.getRoot = function () {
    return this._root;
};

proto.modifyDocument = function ( modificationCallback ) {
    var mutation = this._mutation;
    if ( mutation ) {
        if ( mutation.takeRecords().length ) {
            this._docWasChanged();
        }
        mutation.disconnect();
    }

    this._ignoreAllChanges = true;
    modificationCallback();
    this._ignoreAllChanges = false;

    if ( mutation ) {
        mutation.observe( this._root, {
            childList: true,
            attributes: true,
            characterData: true,
            subtree: true
        });
        this._ignoreChange = false;
    }
};

// --- Events ---

// Subscribing to these events won't automatically add a listener to the
// document node, since these events are fired in a custom manner by the
// editor code.
var customEvents = {
    pathChange: 1, select: 1, input: 1, undoStateChange: 1
};

proto.fireEvent = function ( type, event ) {
    var handlers = this._events[ type ];
    var isFocused, l, obj;
    // UI code, especially modal views, may be monitoring for focus events and
    // immediately removing focus. In certain conditions, this can cause the
    // focus event to fire after the blur event, which can cause an infinite
    // loop. So we detect whether we're actually focused/blurred before firing.
    if ( /^(?:focus|blur)/.test( type ) ) {
        isFocused = this._root === this._doc.activeElement;
        if ( type === 'focus' ) {
            if ( !isFocused || this._isFocused ) {
                return this;
            }
            this._isFocused = true;
        } else {
            if ( isFocused || !this._isFocused ) {
                return this;
            }
            this._isFocused = false;
        }
    }
    if ( handlers ) {
        if ( !event ) {
            event = {};
        }
        if ( event.type !== type ) {
            event.type = type;
        }
        // Clone handlers array, so any handlers added/removed do not affect it.
        handlers = handlers.slice();
        l = handlers.length;
        while ( l-- ) {
            obj = handlers[l];
            try {
                if ( obj.handleEvent ) {
                    obj.handleEvent( event );
                } else {
                    obj.call( this, event );
                }
            } catch ( error ) {
                error.details = 'Squire: fireEvent error. Event type: ' + type;
                this.didError( error );
            }
        }
    }
    return this;
};

proto.destroy = function () {
    var events = this._events;
    var type;

    for ( type in events ) {
        this.removeEventListener( type );
    }
    if ( this._mutation ) {
        this._mutation.disconnect();
    }
    delete this._root.__squire__;

    // Destroy undo stack
    this._undoIndex = -1;
    this._undoStack = [];
    this._undoStackLength = 0;
};

proto.handleEvent = function ( event ) {
    this.fireEvent( event.type, event );
};

proto.addEventListener = function ( type, fn ) {
    var handlers = this._events[ type ];
    var target = this._root;
    if ( !fn ) {
        this.didError({
            name: 'Squire: addEventListener with null or undefined fn',
            message: 'Event type: ' + type
        });
        return this;
    }
    if ( !handlers ) {
        handlers = this._events[ type ] = [];
        if ( !customEvents[ type ] ) {
            if ( type === 'selectionchange' ) {
                target = this._doc;
            }
            target.addEventListener( type, this, true );
        }
    }
    handlers.push( fn );
    return this;
};

proto.removeEventListener = function ( type, fn ) {
    var handlers = this._events[ type ];
    var target = this._root;
    var l;
    if ( handlers ) {
        if ( fn ) {
            l = handlers.length;
            while ( l-- ) {
                if ( handlers[l] === fn ) {
                    handlers.splice( l, 1 );
                }
            }
        } else {
            handlers.length = 0;
        }
        if ( !handlers.length ) {
            delete this._events[ type ];
            if ( !customEvents[ type ] ) {
                if ( type === 'selectionchange' ) {
                    target = this._doc;
                }
                target.removeEventListener( type, this, true );
            }
        }
    }
    return this;
};

// --- Selection and Path ---

proto._createRange =
        function ( range, startOffset, endContainer, endOffset ) {
    if ( range instanceof this._win.Range ) {
        return range.cloneRange();
    }
    var domRange = this._doc.createRange();
    domRange.setStart( range, startOffset );
    if ( endContainer ) {
        domRange.setEnd( endContainer, endOffset );
    } else {
        domRange.setEnd( range, startOffset );
    }
    return domRange;
};

proto.getCursorPosition = function ( range ) {
    if ( ( !range && !( range = this.getSelection() ) ) ||
            !range.getBoundingClientRect ) {
        return null;
    }
    // Get the bounding rect
    var rect = range.getBoundingClientRect();
    var node, parent;
    if ( rect && !rect.top ) {
        this._ignoreChange = true;
        node = this._doc.createElement( 'SPAN' );
        node.textContent = ZWS;
        insertNodeInRange( range, node );
        rect = node.getBoundingClientRect();
        parent = node.parentNode;
        parent.removeChild( node );
        mergeInlines( parent, range );
    }
    return rect;
};

proto._moveCursorTo = function ( toStart ) {
    var root = this._root,
        range = this._createRange( root, toStart ? 0 : root.childNodes.length );
    moveRangeBoundariesDownTree( range );
    this.setSelection( range );
    return this;
};
proto.moveCursorToStart = function () {
    return this._moveCursorTo( true );
};
proto.moveCursorToEnd = function () {
    return this._moveCursorTo( false );
};

var getWindowSelection = function ( self ) {
    return self._win.getSelection() || null;
};

proto.setSelection = function ( range ) {
    if ( range ) {
        this._lastSelection = range;
        // If we're setting selection, that automatically, and synchronously, // triggers a focus event. So just store the selection and mark it as
        // needing restore on focus.
        if ( !this._isFocused ) {
            enableRestoreSelection.call( this );
        } else if ( isAndroid && !this._restoreSelection ) {
            // Android closes the keyboard on removeAllRanges() and doesn't
            // open it again when addRange() is called, sigh.
            // Since Android doesn't trigger a focus event in setSelection(),
            // use a blur/focus dance to work around this by letting the
            // selection be restored on focus.
            // Need to check for !this._restoreSelection to avoid infinite loop
            enableRestoreSelection.call( this );
            this.blur();
            this.focus();
        } else {
            // iOS bug: if you don't focus the iframe before setting the
            // selection, you can end up in a state where you type but the input
            // doesn't get directed into the contenteditable area but is instead
            // lost in a black hole. Very strange.
            if ( isIOS ) {
                this._win.focus();
            }
            var sel = getWindowSelection( this );
            if ( sel ) {
                sel.removeAllRanges();
                sel.addRange( range );
            }
        }
    }
    return this;
};

proto.getSelection = function () {
    var sel = getWindowSelection( this );
    var root = this._root;
    var selection, startContainer, endContainer, node;
    // If not focused, always rely on cached selection; another function may
    // have set it but the DOM is not modified until focus again
    if ( this._isFocused && sel && sel.rangeCount ) {
        selection  = sel.getRangeAt( 0 ).cloneRange();
        startContainer = selection.startContainer;
        endContainer = selection.endContainer;
        // FF can return the selection as being inside an <img>. WTF?
        if ( startContainer && isLeaf( startContainer ) ) {
            selection.setStartBefore( startContainer );
        }
        if ( endContainer && isLeaf( endContainer ) ) {
            selection.setEndBefore( endContainer );
        }
    }
    if ( selection &&
            isOrContains( root, selection.commonAncestorContainer ) ) {
        this._lastSelection = selection;
    } else {
        selection = this._lastSelection;
        node = selection.commonAncestorContainer;
        // Check the editor is in the live document; if not, the range has
        // probably been rewritten by the browser and is bogus
        if ( !isOrContains( node.ownerDocument, node ) ) {
            selection = null;
        }
    }
    if ( !selection ) {
        selection = this._createRange( root.firstChild, 0 );
    }
    return selection;
};

function enableRestoreSelection () {
    this._restoreSelection = true;
}
function disableRestoreSelection () {
    this._restoreSelection = false;
}
function restoreSelection () {
    if ( this._restoreSelection ) {
        this.setSelection( this._lastSelection );
    }
}

proto.getSelectedText = function () {
    var range = this.getSelection();
    if ( !range || range.collapsed ) {
        return '';
    }
    var walker = new TreeWalker(
        range.commonAncestorContainer,
        SHOW_TEXT|SHOW_ELEMENT,
        function ( node ) {
            return isNodeContainedInRange( range, node, true );
        }
    );
    var startContainer = range.startContainer;
    var endContainer = range.endContainer;
    var node = walker.currentNode = startContainer;
    var textContent = '';
    var addedTextInBlock = false;
    var value;

    if ( !walker.filter( node ) ) {
        node = walker.nextNode();
    }

    while ( node ) {
        if ( node.nodeType === TEXT_NODE ) {
            value = node.data;
            if ( value && ( /\S/.test( value ) ) ) {
                if ( node === endContainer ) {
                    value = value.slice( 0, range.endOffset );
                }
                if ( node === startContainer ) {
                    value = value.slice( range.startOffset );
                }
                textContent += value;
                addedTextInBlock = true;
            }
        } else if ( node.nodeName === 'BR' ||
                addedTextInBlock && !isInline( node ) ) {
            textContent += '\n';
            addedTextInBlock = false;
        }
        node = walker.nextNode();
    }

    return textContent;
};

proto.getPath = function () {
    return this._path;
};

// --- Workaround for browsers that can't focus empty text nodes ---

// WebKit bug: https://bugs.webkit.org/show_bug.cgi?id=15256

// Walk down the tree starting at the root and remove any ZWS. If the node only
// contained ZWS space then remove it too. We may want to keep one ZWS node at
// the bottom of the tree so the block can be selected. Define that node as the
// keepNode.
var removeZWS = function ( root, keepNode ) {
    var walker = new TreeWalker( root, SHOW_TEXT, function () {
            return true;
        }, false ),
        parent, node, index;
    while ( node = walker.nextNode() ) {
        while ( ( index = node.data.indexOf( ZWS ) ) > -1  &&
                ( !keepNode || node.parentNode !== keepNode ) ) {
            if ( node.length === 1 ) {
                do {
                    parent = node.parentNode;
                    parent.removeChild( node );
                    node = parent;
                    walker.currentNode = parent;
                } while ( isInline( node ) && !getLength( node ) );
                break;
            } else {
                node.deleteData( index, 1 );
            }
        }
    }
};

proto._didAddZWS = function () {
    this._hasZWS = true;
};
proto._removeZWS = function () {
    if ( !this._hasZWS ) {
        return;
    }
    removeZWS( this._root );
    this._hasZWS = false;
};

// --- Path change events ---

proto._updatePath = function ( range, force ) {
    if ( !range ) {
        return;
    }
    var anchor = range.startContainer,
        focus = range.endContainer,
        newPath;
    if ( force || anchor !== this._lastAnchorNode ||
            focus !== this._lastFocusNode ) {
        this._lastAnchorNode = anchor;
        this._lastFocusNode = focus;
        newPath = ( anchor && focus ) ? ( anchor === focus ) ?
            getPath( focus, this._root ) : '(selection)' : '';
        if ( this._path !== newPath ) {
            this._path = newPath;
            this.fireEvent( 'pathChange', { path: newPath } );
        }
    }
    this.fireEvent( range.collapsed ? 'cursor' : 'select', {
        range: range
    });
};

// selectionchange is fired synchronously in IE when removing current selection
// and when setting new selection; keyup/mouseup may have processing we want
// to do first. Either way, send to next event loop.
proto._updatePathOnEvent = function ( event ) {
    var self = this;
    if ( self._isFocused && !self._willUpdatePath ) {
        self._willUpdatePath = true;
        setTimeout( function () {
            self._willUpdatePath = false;
            self._updatePath( self.getSelection() );
        }, 0 );
    }
};

// --- Focus ---

proto.focus = function () {
    this._root.focus();

    if ( isIE ) {
        this.fireEvent( 'focus' );
    }

    return this;
};

proto.blur = function () {
    this._root.blur();

    if ( isIE ) {
        this.fireEvent( 'blur' );
    }

    return this;
};

// --- Bookmarking ---

var startSelectionId = 'squire-selection-start';
var endSelectionId = 'squire-selection-end';

proto._saveRangeToBookmark = function ( range ) {
    var startNode = this.createElement( 'INPUT', {
            id: startSelectionId,
            type: 'hidden'
        }),
        endNode = this.createElement( 'INPUT', {
            id: endSelectionId,
            type: 'hidden'
        }),
        temp;

    insertNodeInRange( range, startNode );
    range.collapse( false );
    insertNodeInRange( range, endNode );

    // In a collapsed range, the start is sometimes inserted after the end!
    if ( startNode.compareDocumentPosition( endNode ) &
            DOCUMENT_POSITION_PRECEDING ) {
        startNode.id = endSelectionId;
        endNode.id = startSelectionId;
        temp = startNode;
        startNode = endNode;
        endNode = temp;
    }

    range.setStartAfter( startNode );
    range.setEndBefore( endNode );
};

proto._getRangeAndRemoveBookmark = function ( range ) {
    var root = this._root,
        start = root.querySelector( '#' + startSelectionId ),
        end = root.querySelector( '#' + endSelectionId );

    if ( start && end ) {
        var startContainer = start.parentNode,
            endContainer = end.parentNode,
            startOffset = indexOf.call( startContainer.childNodes, start ),
            endOffset = indexOf.call( endContainer.childNodes, end );

        if ( startContainer === endContainer ) {
            endOffset -= 1;
        }

        detach( start );
        detach( end );

        if ( !range ) {
            range = this._doc.createRange();
        }
        range.setStart( startContainer, startOffset );
        range.setEnd( endContainer, endOffset );

        // Merge any text nodes we split
        mergeInlines( startContainer, range );
        if ( startContainer !== endContainer ) {
            mergeInlines( endContainer, range );
        }

        // If we didn't split a text node, we should move into any adjacent
        // text node to current selection point
        if ( range.collapsed ) {
            startContainer = range.startContainer;
            if ( startContainer.nodeType === TEXT_NODE ) {
                endContainer = startContainer.childNodes[ range.startOffset ];
                if ( !endContainer || endContainer.nodeType !== TEXT_NODE ) {
                    endContainer =
                        startContainer.childNodes[ range.startOffset - 1 ];
                }
                if ( endContainer && endContainer.nodeType === TEXT_NODE ) {
                    range.setStart( endContainer, 0 );
                    range.collapse( true );
                }
            }
        }
    }
    return range || null;
};

// --- Undo ---

proto._keyUpDetectChange = function ( event ) {
    var code = event.keyCode;
    // Presume document was changed if:
    // 1. A modifier key (other than shift) wasn't held down
    // 2. The key pressed is not in range 16<=x<=20 (control keys)
    // 3. The key pressed is not in range 33<=x<=45 (navigation keys)
    if ( !event.ctrlKey && !event.metaKey && !event.altKey &&
            ( code < 16 || code > 20 ) &&
            ( code < 33 || code > 45 ) ) {
        this._docWasChanged();
    }
};

proto._docWasChanged = function () {
    if ( canWeakMap ) {
        nodeCategoryCache = new WeakMap();
    }
    if ( this._ignoreAllChanges ) {
        return;
    }

    if ( canObserveMutations && this._ignoreChange ) {
        this._ignoreChange = false;
        return;
    }
    if ( this._isInUndoState ) {
        this._isInUndoState = false;
        this.fireEvent( 'undoStateChange', {
            canUndo: true,
            canRedo: false
        });
    }
    this.fireEvent( 'input' );
};

// Leaves bookmark
proto._recordUndoState = function ( range, replace ) {
    // Don't record if we're already in an undo state
    if ( !this._isInUndoState|| replace ) {
        // Advance pointer to new position
        var undoIndex = this._undoIndex;
        var undoStack = this._undoStack;
        var undoConfig = this._config.undo;
        var undoThreshold = undoConfig.documentSizeThreshold;
        var undoLimit = undoConfig.undoLimit;
        var html;

        if ( !replace ) {
            undoIndex += 1;
        }

        // Truncate stack if longer (i.e. if has been previously undone)
        if ( undoIndex < this._undoStackLength ) {
            undoStack.length = this._undoStackLength = undoIndex;
        }

        // Get data
        if ( range ) {
            this._saveRangeToBookmark( range );
        }
        html = this._getHTML();

        // If this document is above the configured size threshold,
        // limit the number of saved undo states.
        // Threshold is in bytes, JS uses 2 bytes per character
        if ( undoThreshold > -1 && html.length * 2 > undoThreshold ) {
            if ( undoLimit > -1 && undoIndex > undoLimit ) {
                undoStack.splice( 0, undoIndex - undoLimit );
                undoIndex = undoLimit;
                this._undoStackLength = undoLimit;
            }
        }

        // Save data
        undoStack[ undoIndex ] = html;
        this._undoIndex = undoIndex;
        this._undoStackLength += 1;
        this._isInUndoState = true;
    }
};

proto.saveUndoState = function ( range ) {
    if ( range === undefined ) {
        range = this.getSelection();
    }
    this._recordUndoState( range, this._isInUndoState );
    this._getRangeAndRemoveBookmark( range );

    return this;
};

proto.undo = function () {
    // Sanity check: must not be at beginning of the history stack
    if ( this._undoIndex !== 0 || !this._isInUndoState ) {
        // Make sure any changes since last checkpoint are saved.
        this._recordUndoState( this.getSelection(), false );

        this._undoIndex -= 1;
        this._setHTML( this._undoStack[ this._undoIndex ] );
        var range = this._getRangeAndRemoveBookmark();
        if ( range ) {
            this.setSelection( range );
        }
        this._isInUndoState = true;
        this.fireEvent( 'undoStateChange', {
            canUndo: this._undoIndex !== 0,
            canRedo: true
        });
        this.fireEvent( 'input' );
    }
    return this;
};

proto.redo = function () {
    // Sanity check: must not be at end of stack and must be in an undo
    // state.
    var undoIndex = this._undoIndex,
        undoStackLength = this._undoStackLength;
    if ( undoIndex + 1 < undoStackLength && this._isInUndoState ) {
        this._undoIndex += 1;
        this._setHTML( this._undoStack[ this._undoIndex ] );
        var range = this._getRangeAndRemoveBookmark();
        if ( range ) {
            this.setSelection( range );
        }
        this.fireEvent( 'undoStateChange', {
            canUndo: true,
            canRedo: undoIndex + 2 < undoStackLength
        });
        this.fireEvent( 'input' );
    }
    return this;
};

// --- Inline formatting ---

// Looks for matching tag and attributes, so won't work
// if <strong> instead of <b> etc.
proto.hasFormat = function ( tag, attributes, range ) {
    // 1. Normalise the arguments and get selection
    tag = tag.toUpperCase();
    if ( !attributes ) { attributes = {}; }
    if ( !range && !( range = this.getSelection() ) ) {
        return false;
    }

    // Sanitize range to prevent weird IE artifacts
    if ( !range.collapsed &&
            range.startContainer.nodeType === TEXT_NODE &&
            range.startOffset === range.startContainer.length &&
            range.startContainer.nextSibling ) {
        range.setStartBefore( range.startContainer.nextSibling );
    }
    if ( !range.collapsed &&
            range.endContainer.nodeType === TEXT_NODE &&
            range.endOffset === 0 &&
            range.endContainer.previousSibling ) {
        range.setEndAfter( range.endContainer.previousSibling );
    }

    // If the common ancestor is inside the tag we require, we definitely
    // have the format.
    var root = this._root;
    var common = range.commonAncestorContainer;
    var walker, node;
    if ( getNearest( common, root, tag, attributes ) ) {
        return true;
    }

    // If common ancestor is a text node and doesn't have the format, we
    // definitely don't have it.
    if ( common.nodeType === TEXT_NODE ) {
        return false;
    }

    // Otherwise, check each text node at least partially contained within
    // the selection and make sure all of them have the format we want.
    walker = new TreeWalker( common, SHOW_TEXT, function ( node ) {
        return isNodeContainedInRange( range, node, true );
    }, false );

    var seenNode = false;
    while ( node = walker.nextNode() ) {
        if ( !getNearest( node, root, tag, attributes ) ) {
            return false;
        }
        seenNode = true;
    }

    return seenNode;
};

// Extracts the font-family and font-size (if any) of the element
// holding the cursor. If there's a selection, returns an empty object.
proto.getFontInfo = function ( range ) {
    var fontInfo = {
        color: undefined,
        backgroundColor: undefined,
        family: undefined,
        size: undefined
    };
    var seenAttributes = 0;
    var element, style, attr;

    if ( !range && !( range = this.getSelection() ) ) {
        return fontInfo;
    }

    element = range.commonAncestorContainer;
    if ( range.collapsed || element.nodeType === TEXT_NODE ) {
        if ( element.nodeType === TEXT_NODE ) {
            element = element.parentNode;
        }
        while ( seenAttributes < 4 && element ) {
            if ( style = element.style ) {
                if ( !fontInfo.color && ( attr = style.color ) ) {
                    fontInfo.color = attr;
                    seenAttributes += 1;
                }
                if ( !fontInfo.backgroundColor &&
                        ( attr = style.backgroundColor ) ) {
                    fontInfo.backgroundColor = attr;
                    seenAttributes += 1;
                }
                if ( !fontInfo.family && ( attr = style.fontFamily ) ) {
                    fontInfo.family = attr;
                    seenAttributes += 1;
                }
                if ( !fontInfo.size && ( attr = style.fontSize ) ) {
                    fontInfo.size = attr;
                    seenAttributes += 1;
                }
            }
            element = element.parentNode;
        }
    }
    return fontInfo;
};

proto._addFormat = function ( tag, attributes, range ) {
    // If the range is collapsed we simply insert the node by wrapping
    // it round the range and focus it.
    var root = this._root;
    var el, walker, startContainer, endContainer, startOffset, endOffset,
        node, needsFormat, block;

    if ( range.collapsed ) {
        el = fixCursor( this.createElement( tag, attributes ), root );
        insertNodeInRange( range, el );
        range.setStart( el.firstChild, el.firstChild.length );
        range.collapse( true );

        // Clean up any previous formats that may have been set on this block
        // that are unused.
        block = el;
        while ( isInline( block ) ) {
            block = block.parentNode;
        }
        removeZWS( block, el );
    }
    // Otherwise we find all the textnodes in the range (splitting
    // partially selected nodes) and if they're not already formatted
    // correctly we wrap them in the appropriate tag.
    else {
        // Create an iterator to walk over all the text nodes under this
        // ancestor which are in the range and not already formatted
        // correctly.
        //
        // In Blink/WebKit, empty blocks may have no text nodes, just a <br>.
        // Therefore we wrap this in the tag as well, as this will then cause it
        // to apply when the user types something in the block, which is
        // presumably what was intended.
        //
        // IMG tags are included because we may want to create a link around
        // them, and adding other styles is harmless.
        walker = new TreeWalker(
            range.commonAncestorContainer,
            SHOW_TEXT|SHOW_ELEMENT,
            function ( node ) {
                return ( node.nodeType === TEXT_NODE ||
                        node.nodeName === 'BR' ||
                        node.nodeName === 'IMG'
                    ) && isNodeContainedInRange( range, node, true );
            },
            false
        );

        // Start at the beginning node of the range and iterate through
        // all the nodes in the range that need formatting.
        startContainer = range.startContainer;
        startOffset = range.startOffset;
        endContainer = range.endContainer;
        endOffset = range.endOffset;

        // Make sure we start with a valid node.
        walker.currentNode = startContainer;
        if ( !walker.filter( startContainer ) ) {
            startContainer = walker.nextNode();
            startOffset = 0;
        }

        // If there are no interesting nodes in the selection, abort
        if ( !startContainer ) {
            return range;
        }

        do {
            node = walker.currentNode;
            needsFormat = !getNearest( node, root, tag, attributes );
            if ( needsFormat ) {
                // <br> can never be a container node, so must have a text node
                // if node == (end|start)Container
                if ( node === endContainer && node.length > endOffset ) {
                    node.splitText( endOffset );
                }
                if ( node === startContainer && startOffset ) {
                    node = node.splitText( startOffset );
                    if ( endContainer === startContainer ) {
                        endContainer = node;
                        endOffset -= startOffset;
                    }
                    startContainer = node;
                    startOffset = 0;
                }
                el = this.createElement( tag, attributes );
                replaceWith( node, el );
                el.appendChild( node );
            }
        } while ( walker.nextNode() );

        // If we don't finish inside a text node, offset may have changed.
        if ( endContainer.nodeType !== TEXT_NODE ) {
            if ( node.nodeType === TEXT_NODE ) {
                endContainer = node;
                endOffset = node.length;
            } else {
                // If <br>, we must have just wrapped it, so it must have only
                // one child
                endContainer = node.parentNode;
                endOffset = 1;
            }
        }

        // Now set the selection to as it was before
        range = this._createRange(
            startContainer, startOffset, endContainer, endOffset );
    }
    return range;
};

proto._removeFormat = function ( tag, attributes, range, partial ) {
    // Add bookmark
    this._saveRangeToBookmark( range );

    // We need a node in the selection to break the surrounding
    // formatted text.
    var doc = this._doc,
        fixer;
    if ( range.collapsed ) {
        if ( cantFocusEmptyTextNodes ) {
            fixer = doc.createTextNode( ZWS );
            this._didAddZWS();
        } else {
            fixer = doc.createTextNode( '' );
        }
        insertNodeInRange( range, fixer );
    }

    // Find block-level ancestor of selection
    var root = range.commonAncestorContainer;
    while ( isInline( root ) ) {
        root = root.parentNode;
    }

    // Find text nodes inside formatTags that are not in selection and
    // add an extra tag with the same formatting.
    var startContainer = range.startContainer,
        startOffset = range.startOffset,
        endContainer = range.endContainer,
        endOffset = range.endOffset,
        toWrap = [],
        examineNode = function ( node, exemplar ) {
            // If the node is completely contained by the range then
            // we're going to remove all formatting so ignore it.
            if ( isNodeContainedInRange( range, node, false ) ) {
                return;
            }

            var isText = ( node.nodeType === TEXT_NODE ),
                child, next;

            // If not at least partially contained, wrap entire contents
            // in a clone of the tag we're removing and we're done.
            if ( !isNodeContainedInRange( range, node, true ) ) {
                // Ignore bookmarks and empty text nodes
                if ( node.nodeName !== 'INPUT' &&
                        ( !isText || node.data ) ) {
                    toWrap.push([ exemplar, node ]);
                }
                return;
            }

            // Split any partially selected text nodes.
            if ( isText ) {
                if ( node === endContainer && endOffset !== node.length ) {
                    toWrap.push([ exemplar, node.splitText( endOffset ) ]);
                }
                if ( node === startContainer && startOffset ) {
                    node.splitText( startOffset );
                    toWrap.push([ exemplar, node ]);
                }
            }
            // If not a text node, recurse onto all children.
            // Beware, the tree may be rewritten with each call
            // to examineNode, hence find the next sibling first.
            else {
                for ( child = node.firstChild; child; child = next ) {
                    next = child.nextSibling;
                    examineNode( child, exemplar );
                }
            }
        },
        formatTags = Array.prototype.filter.call(
            root.getElementsByTagName( tag ), function ( el ) {
                return isNodeContainedInRange( range, el, true ) &&
                    hasTagAttributes( el, tag, attributes );
            }
        );

    if ( !partial ) {
        formatTags.forEach( function ( node ) {
            examineNode( node, node );
        });
    }

    // Now wrap unselected nodes in the tag
    toWrap.forEach( function ( item ) {
        // [ exemplar, node ] tuple
        var el = item[0].cloneNode( false ),
            node = item[1];
        replaceWith( node, el );
        el.appendChild( node );
    });
    // and remove old formatting tags.
    formatTags.forEach( function ( el ) {
        replaceWith( el, empty( el ) );
    });

    // Merge adjacent inlines:
    this._getRangeAndRemoveBookmark( range );
    if ( fixer ) {
        range.collapse( false );
    }
    mergeInlines( root, range );

    return range;
};

proto.changeFormat = function ( add, remove, range, partial ) {
    // Normalise the arguments and get selection
    if ( !range && !( range = this.getSelection() ) ) {
        return this;
    }

    // Save undo checkpoint
    this.saveUndoState( range );

    if ( remove ) {
        range = this._removeFormat( remove.tag.toUpperCase(),
            remove.attributes || {}, range, partial );
    }
    if ( add ) {
        range = this._addFormat( add.tag.toUpperCase(),
            add.attributes || {}, range );
    }

    this.setSelection( range );
    this._updatePath( range, true );

    // We're not still in an undo state
    if ( !canObserveMutations ) {
        this._docWasChanged();
    }

    return this;
};

// --- Block formatting ---

var tagAfterSplit = {
    DT:  'DD',
    DD:  'DT',
    LI:  'LI',
    PRE: 'PRE'
};

var splitBlock = function ( self, block, node, offset ) {
    var splitTag = tagAfterSplit[ block.nodeName ],
        splitProperties = null,
        nodeAfterSplit = split( node, offset, block.parentNode, self._root ),
        config = self._config;

    if ( !splitTag ) {
        splitTag = config.blockTag;
        splitProperties = config.blockAttributes;
    }

    // Make sure the new node is the correct type.
    if ( !hasTagAttributes( nodeAfterSplit, splitTag, splitProperties ) ) {
        block = createElement( nodeAfterSplit.ownerDocument,
            splitTag, splitProperties );
        if ( nodeAfterSplit.dir ) {
            block.dir = nodeAfterSplit.dir;
        }
        replaceWith( nodeAfterSplit, block );
        block.appendChild( empty( nodeAfterSplit ) );
        nodeAfterSplit = block;
    }
    return nodeAfterSplit;
};

proto.forEachBlock = function ( fn, mutates, range ) {
    if ( !range && !( range = this.getSelection() ) ) {
        return this;
    }

    // Save undo checkpoint
    if ( mutates ) {
        this.saveUndoState( range );
    }

    var root = this._root;
    var start = getStartBlockOfRange( range, root );
    var end = getEndBlockOfRange( range, root );
    if ( start && end ) {
        do {
            if ( fn( start ) || start === end ) { break; }
        } while ( start = getNextBlock( start, root ) );
    }

    if ( mutates ) {
        this.setSelection( range );

        // Path may have changed
        this._updatePath( range, true );

        // We're not still in an undo state
        if ( !canObserveMutations ) {
            this._docWasChanged();
        }
    }
    return this;
};

proto.modifyBlocks = function ( modify, range ) {
    if ( !range && !( range = this.getSelection() ) ) {
        return this;
    }

    // 1. Save undo checkpoint and bookmark selection
    this._recordUndoState( range, this._isInUndoState );

    var root = this._root;
    var frag;

    // 2. Expand range to block boundaries
    expandRangeToBlockBoundaries( range, root );

    // 3. Remove range.
    moveRangeBoundariesUpTree( range, root, root, root );
    frag = extractContentsOfRange( range, root, root );

    // 4. Modify tree of fragment and reinsert.
    insertNodeInRange( range, modify.call( this, frag ) );

    // 5. Merge containers at edges
    if ( range.endOffset < range.endContainer.childNodes.length ) {
        mergeContainers( range.endContainer.childNodes[ range.endOffset ], root );
    }
    mergeContainers( range.startContainer.childNodes[ range.startOffset ], root );

    // 6. Restore selection
    this._getRangeAndRemoveBookmark( range );
    this.setSelection( range );
    this._updatePath( range, true );

    // 7. We're not still in an undo state
    if ( !canObserveMutations ) {
        this._docWasChanged();
    }

    return this;
};

var increaseBlockQuoteLevel = function ( frag ) {
    return this.createElement( 'BLOCKQUOTE',
        this._config.tagAttributes.blockquote, [
            frag
        ]);
};

var decreaseBlockQuoteLevel = function ( frag ) {
    var root = this._root;
    var blockquotes = frag.querySelectorAll( 'blockquote' );
    Array.prototype.filter.call( blockquotes, function ( el ) {
        return !getNearest( el.parentNode, root, 'BLOCKQUOTE' );
    }).forEach( function ( el ) {
        replaceWith( el, empty( el ) );
    });
    return frag;
};

var removeBlockQuote = function (/* frag */) {
    return this.createDefaultBlock([
        this.createElement( 'INPUT', {
            id: startSelectionId,
            type: 'hidden'
        }),
        this.createElement( 'INPUT', {
            id: endSelectionId,
            type: 'hidden'
        })
    ]);
};

var makeList = function ( self, frag, type ) {
    var walker = getBlockWalker( frag, self._root ),
        node, tag, prev, newLi,
        tagAttributes = self._config.tagAttributes,
        listAttrs = tagAttributes[ type.toLowerCase() ],
        listItemAttrs = tagAttributes.li;

    while ( node = walker.nextNode() ) {
        if ( node.parentNode.nodeName === 'LI' ) {
            node = node.parentNode;
            walker.currentNode = node.lastChild;
        }
        if ( node.nodeName !== 'LI' ) {
            newLi = self.createElement( 'LI', listItemAttrs );
            if ( node.dir ) {
                newLi.dir = node.dir;
            }

            // Have we replaced the previous block with a new <ul>/<ol>?
            if ( ( prev = node.previousSibling ) && prev.nodeName === type ) {
                prev.appendChild( newLi );
                detach( node );
            }
            // Otherwise, replace this block with the <ul>/<ol>
            else {
                replaceWith(
                    node,
                    self.createElement( type, listAttrs, [
                        newLi
                    ])
                );
            }
            newLi.appendChild( empty( node ) );
            walker.currentNode = newLi;
        } else {
            node = node.parentNode;
            tag = node.nodeName;
            if ( tag !== type && ( /^[OU]L$/.test( tag ) ) ) {
                replaceWith( node,
                    self.createElement( type, listAttrs, [ empty( node ) ] )
                );
            }
        }
    }
};

var makeUnorderedList = function ( frag ) {
    makeList( this, frag, 'UL' );
    return frag;
};

var makeOrderedList = function ( frag ) {
    makeList( this, frag, 'OL' );
    return frag;
};

var removeList = function ( frag ) {
    var lists = frag.querySelectorAll( 'UL, OL' ),
        items =  frag.querySelectorAll( 'LI' ),
        root = this._root,
        i, l, list, listFrag, item;
    for ( i = 0, l = lists.length; i < l; i += 1 ) {
        list = lists[i];
        listFrag = empty( list );
        fixContainer( listFrag, root );
        replaceWith( list, listFrag );
    }

    for ( i = 0, l = items.length; i < l; i += 1 ) {
        item = items[i];
        if ( isBlock( item ) ) {
            replaceWith( item,
                this.createDefaultBlock([ empty( item ) ])
            );
        } else {
            fixContainer( item, root );
            replaceWith( item, empty( item ) );
        }
    }
    return frag;
};

var getListSelection = function ( range, root ) {
    // Get start+end li in single common ancestor
    var list = range.commonAncestorContainer;
    var startLi = range.startContainer;
    var endLi = range.endContainer;
    while ( list && list !== root && !/^[OU]L$/.test( list.nodeName ) ) {
        list = list.parentNode;
    }
    if ( !list || list === root ) {
        return null;
    }
    if ( startLi === list ) {
        startLi = startLi.childNodes[ range.startOffset ];
    }
    if ( endLi === list ) {
        endLi = endLi.childNodes[ range.endOffset ];
    }
    while ( startLi && startLi.parentNode !== list ) {
        startLi = startLi.parentNode;
    }
    while ( endLi && endLi.parentNode !== list ) {
        endLi = endLi.parentNode;
    }
    return [ list, startLi, endLi ];
};

proto.increaseListLevel = function ( range ) {
    if ( !range && !( range = this.getSelection() ) ) {
        return this.focus();
    }

    var root = this._root;
    var listSelection = getListSelection( range, root );
    if ( !listSelection ) {
        return this.focus();
    }

    var list = listSelection[0];
    var startLi = listSelection[1];
    var endLi = listSelection[2];
    if ( !startLi || startLi === list.firstChild ) {
        return this.focus();
    }

    // Save undo checkpoint and bookmark selection
    this._recordUndoState( range, this._isInUndoState );

    // Increase list depth
    var type = list.nodeName;
    var newParent = startLi.previousSibling;
    var listAttrs, next;
    if ( newParent.nodeName !== type ) {
        listAttrs = this._config.tagAttributes[ type.toLowerCase() ];
        newParent = this.createElement( type, listAttrs );
        list.insertBefore( newParent, startLi );
    }
    do {
        next = startLi === endLi ? null : startLi.nextSibling;
        newParent.appendChild( startLi );
    } while ( ( startLi = next ) );
    next = newParent.nextSibling;
    if ( next ) {
        mergeContainers( next, root );
    }

    // Restore selection
    this._getRangeAndRemoveBookmark( range );
    this.setSelection( range );
    this._updatePath( range, true );

    // We're not still in an undo state
    if ( !canObserveMutations ) {
        this._docWasChanged();
    }

    return this.focus();
};

proto.decreaseListLevel = function ( range ) {
    if ( !range && !( range = this.getSelection() ) ) {
        return this.focus();
    }

    var root = this._root;
    var listSelection = getListSelection( range, root );
    if ( !listSelection ) {
        return this.focus();
    }

    var list = listSelection[0];
    var startLi = listSelection[1];
    var endLi = listSelection[2];
    if ( !startLi ) {
        startLi = list.firstChild;
    }
    if ( !endLi ) {
        endLi = list.lastChild;
    }

    // Save undo checkpoint and bookmark selection
    this._recordUndoState( range, this._isInUndoState );

    // Find the new parent list node
    var newParent = list.parentNode;
    var next;

    // Split list if necesary
    var insertBefore = !endLi.nextSibling ?
        list.nextSibling :
        split( list, endLi.nextSibling, newParent, root );

    if ( newParent !== root && newParent.nodeName === 'LI' ) {
        newParent = newParent.parentNode;
        while ( insertBefore ) {
            next = insertBefore.nextSibling;
            endLi.appendChild( insertBefore );
            insertBefore = next;
        }
        insertBefore = list.parentNode.nextSibling;
    }

    var makeNotList = !/^[OU]L$/.test( newParent.nodeName );
    do {
        next = startLi === endLi ? null : startLi.nextSibling;
        list.removeChild( startLi );
        if ( makeNotList && startLi.nodeName === 'LI' ) {
            startLi = this.createDefaultBlock([ empty( startLi ) ]);
        }
        newParent.insertBefore( startLi, insertBefore );
    } while ( ( startLi = next ) );

    if ( !list.firstChild ) {
        detach( list );
    }

    if ( insertBefore ) {
        mergeContainers( insertBefore, root );
    }

    // Restore selection
    this._getRangeAndRemoveBookmark( range );
    this.setSelection( range );
    this._updatePath( range, true );

    // We're not still in an undo state
    if ( !canObserveMutations ) {
        this._docWasChanged();
    }

    return this.focus();
};

proto._ensureBottomLine = function () {
    var root = this._root;
    var last = root.lastElementChild;
    if ( !last ||
            last.nodeName !== this._config.blockTag || !isBlock( last ) ) {
        root.appendChild( this.createDefaultBlock() );
    }
};

// --- Keyboard interaction ---

proto.setKeyHandler = function ( key, fn ) {
    this._keyHandlers[ key ] = fn;
    return this;
};

// --- Get/Set data ---

proto._getHTML = function () {
    return this._root.innerHTML;
};

proto._setHTML = function ( html ) {
    var root = this._root;
    var node = root;
    node.innerHTML = html;
    do {
        fixCursor( node, root );
    } while ( node = getNextBlock( node, root ) );
    this._ignoreChange = true;
};

proto.getHTML = function ( withBookMark ) {
    var brs = [],
        root, node, fixer, html, l, range;
    if ( withBookMark && ( range = this.getSelection() ) ) {
        this._saveRangeToBookmark( range );
    }
    if ( useTextFixer ) {
        root = this._root;
        node = root;
        while ( node = getNextBlock( node, root ) ) {
            if ( !node.textContent && !node.querySelector( 'BR' ) ) {
                fixer = this.createElement( 'BR' );
                node.appendChild( fixer );
                brs.push( fixer );
            }
        }
    }
    html = this._getHTML().replace( /\u200B/g, '' );
    if ( useTextFixer ) {
        l = brs.length;
        while ( l-- ) {
            detach( brs[l] );
        }
    }
    if ( range ) {
        this._getRangeAndRemoveBookmark( range );
    }
    return html;
};

proto.setHTML = function ( html ) {
    var config = this._config;
    var sanitizeToDOMFragment = config.isSetHTMLSanitized ?
            config.sanitizeToDOMFragment : null;
    var root = this._root;
    var div, frag, child;

    // Parse HTML into DOM tree
    if ( typeof sanitizeToDOMFragment === 'function' ) {
        frag = sanitizeToDOMFragment( html, false, this );
    } else {
        div = this.createElement( 'DIV' );
        div.innerHTML = html;
        frag = this._doc.createDocumentFragment();
        frag.appendChild( empty( div ) );
    }

    cleanTree( frag );
    cleanupBRs( frag, root, false );

    fixContainer( frag, root );

    // Fix cursor
    var node = frag;
    while ( node = getNextBlock( node, root ) ) {
        fixCursor( node, root );
    }

    // Don't fire an input event
    this._ignoreChange = true;

    // Remove existing root children
    while ( child = root.lastChild ) {
        root.removeChild( child );
    }

    // And insert new content
    root.appendChild( frag );
    fixCursor( root, root );

    // Reset the undo stack
    this._undoIndex = -1;
    this._undoStack.length = 0;
    this._undoStackLength = 0;
    this._isInUndoState = false;

    // Record undo state
    var range = this._getRangeAndRemoveBookmark() ||
        this._createRange( root.firstChild, 0 );
    this.saveUndoState( range );
    // IE will also set focus when selecting text so don't use
    // setSelection. Instead, just store it in lastSelection, so if
    // anything calls getSelection before first focus, we have a range
    // to return.
    this._lastSelection = range;
    enableRestoreSelection.call( this );
    this._updatePath( range, true );

    return this;
};

proto.insertElement = function ( el, range ) {
    if ( !range ) {
        range = this.getSelection();
    }
    range.collapse( true );
    if ( isInline( el ) ) {
        insertNodeInRange( range, el );
        range.setStartAfter( el );
    } else {
        // Get containing block node.
        var root = this._root;
        var splitNode = getStartBlockOfRange( range, root ) || root;
        var parent, nodeAfterSplit;
        // While at end of container node, move up DOM tree.
        while ( splitNode !== root && !splitNode.nextSibling ) {
            splitNode = splitNode.parentNode;
        }
        // If in the middle of a container node, split up to root.
        if ( splitNode !== root ) {
            parent = splitNode.parentNode;
            nodeAfterSplit = split( parent, splitNode.nextSibling, root, root );
        }
        if ( nodeAfterSplit ) {
            root.insertBefore( el, nodeAfterSplit );
        } else {
            root.appendChild( el );
            // Insert blank line below block.
            nodeAfterSplit = this.createDefaultBlock();
            root.appendChild( nodeAfterSplit );
        }
        range.setStart( nodeAfterSplit, 0 );
        range.setEnd( nodeAfterSplit, 0 );
        moveRangeBoundariesDownTree( range );
    }
    this.focus();
    this.setSelection( range );
    this._updatePath( range );

    if ( !canObserveMutations ) {
        this._docWasChanged();
    }

    return this;
};

proto.insertImage = function ( src, attributes ) {
    var img = this.createElement( 'IMG', mergeObjects({
        src: src
    }, attributes, true ));
    this.insertElement( img );
    return img;
};

var linkRegExp = /\b((?:(?:ht|f)tps?:\/\/|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,}\/)(?:[^\s()<>]+|\([^\s()<>]+\))+(?:\((?:[^\s()<>]+|(?:\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))|([\w\-.%+]+@(?:[\w\-]+\.)+[A-Z]{2,}\b)/i;

var addLinks = function ( frag, root, self ) {
    var doc = frag.ownerDocument,
        walker = new TreeWalker( frag, SHOW_TEXT,
                function ( node ) {
            return !getNearest( node, root, 'A' );
        }, false ),
        defaultAttributes = self._config.tagAttributes.a,
        node, data, parent, match, index, endIndex, child;
    while ( node = walker.nextNode() ) {
        data = node.data;
        parent = node.parentNode;
        while ( match = linkRegExp.exec( data ) ) {
            index = match.index;
            endIndex = index + match[0].length;
            if ( index ) {
                child = doc.createTextNode( data.slice( 0, index ) );
                parent.insertBefore( child, node );
            }
            child = self.createElement( 'A', mergeObjects({
                href: match[1] ?
                    /^(?:ht|f)tps?:/.test( match[1] ) ?
                        match[1] :
                        'http://' + match[1] :
                    'mailto:' + match[2]
            }, defaultAttributes, false ));
            child.textContent = data.slice( index, endIndex );
            parent.insertBefore( child, node );
            node.data = data = data.slice( endIndex );
        }
    }
};

// Insert HTML at the cursor location. If the selection is not collapsed
// insertTreeFragmentIntoRange will delete the selection so that it is replaced
// by the html being inserted.
proto.insertHTML = function ( html, isPaste ) {
    var config = this._config;
    var sanitizeToDOMFragment = config.isInsertedHTMLSanitized ?
            config.sanitizeToDOMFragment : null;
    var range = this.getSelection();
    var doc = this._doc;
    var startFragmentIndex, endFragmentIndex;
    var div, frag, root, node, event;

    // Edge doesn't just copy the fragment, but includes the surrounding guff
    // including the full <head> of the page. Need to strip this out. If
    // available use DOMPurify to parse and sanitise.
    if ( typeof sanitizeToDOMFragment === 'function' ) {
        frag = sanitizeToDOMFragment( html, isPaste, this );
    } else {
        if ( isPaste ) {
            startFragmentIndex = html.indexOf( '<!--StartFragment-->' );
            endFragmentIndex = html.lastIndexOf( '<!--EndFragment-->' );
            if ( startFragmentIndex > -1 && endFragmentIndex > -1 ) {
                html = html.slice( startFragmentIndex + 20, endFragmentIndex );
            }
        }
        // Wrap with <tr> if html contains dangling <td> tags
        if ( /<\/td>((?!<\/tr>)[\s\S])*$/i.test( html ) ) {
            html = '<TR>' + html + '</TR>';
        }
        // Wrap with <table> if html contains dangling <tr> tags
        if ( /<\/tr>((?!<\/table>)[\s\S])*$/i.test( html ) ) {
            html = '<TABLE>' + html + '</TABLE>';
        }
        // Parse HTML into DOM tree
        div = this.createElement( 'DIV' );
        div.innerHTML = html;
        frag = doc.createDocumentFragment();
        frag.appendChild( empty( div ) );
    }

    // Record undo checkpoint
    this.saveUndoState( range );

    try {
        root = this._root;
        node = frag;
        event = {
            fragment: frag,
            preventDefault: function () {
                this.defaultPrevented = true;
            },
            defaultPrevented: false
        };

        addLinks( frag, frag, this );
        cleanTree( frag );
        cleanupBRs( frag, root, false );
        removeEmptyInlines( frag );
        frag.normalize();

        while ( node = getNextBlock( node, frag ) ) {
            fixCursor( node, root );
        }

        if ( isPaste ) {
            this.fireEvent( 'willPaste', event );
        }

        if ( !event.defaultPrevented ) {
            insertTreeFragmentIntoRange( range, event.fragment, root );
            if ( !canObserveMutations ) {
                this._docWasChanged();
            }
            range.collapse( false );
            this._ensureBottomLine();
        }

        this.setSelection( range );
        this._updatePath( range, true );
        // Safari sometimes loses focus after paste. Weird.
        if ( isPaste ) {
            this.focus();
        }
    } catch ( error ) {
        this.didError( error );
    }
    return this;
};

var escapeHTMLFragement = function ( text ) {
    return text.split( '&' ).join( '&amp;' )
               .split( '<' ).join( '&lt;'  )
               .split( '>' ).join( '&gt;'  )
               .split( '"' ).join( '&quot;'  );
};

proto.insertPlainText = function ( plainText, isPaste ) {
    var lines = plainText.split( '\n' );
    var config = this._config;
    var tag = config.blockTag;
    var attributes = config.blockAttributes;
    var closeBlock  = '</' + tag + '>';
    var openBlock = '<' + tag;
    var attr, i, l, line;

    for ( attr in attributes ) {
        openBlock += ' ' + attr + '="' +
            escapeHTMLFragement( attributes[ attr ] ) +
        '"';
    }
    openBlock += '>';

    for ( i = 0, l = lines.length; i < l; i += 1 ) {
        line = lines[i];
        line = escapeHTMLFragement( line ).replace( / (?= )/g, '&nbsp;' );
        // Wrap each line in <div></div>
        lines[i] = openBlock + ( line || '<BR>' ) + closeBlock;
    }
    return this.insertHTML( lines.join( '' ), isPaste );
};

// --- Formatting ---

var command = function ( method, arg, arg2 ) {
    return function () {
        this[ method ]( arg, arg2 );
        return this.focus();
    };
};

proto.addStyles = function ( styles ) {
    if ( styles ) {
        var head = this._doc.documentElement.firstChild,
            style = this.createElement( 'STYLE', {
                type: 'text/css'
            });
        style.appendChild( this._doc.createTextNode( styles ) );
        head.appendChild( style );
    }
    return this;
};

proto.bold = command( 'changeFormat', { tag: 'B' } );
proto.italic = command( 'changeFormat', { tag: 'I' } );
proto.underline = command( 'changeFormat', { tag: 'U' } );
proto.strikethrough = command( 'changeFormat', { tag: 'S' } );
proto.subscript = command( 'changeFormat', { tag: 'SUB' }, { tag: 'SUP' } );
proto.superscript = command( 'changeFormat', { tag: 'SUP' }, { tag: 'SUB' } );

proto.removeBold = command( 'changeFormat', null, { tag: 'B' } );
proto.removeItalic = command( 'changeFormat', null, { tag: 'I' } );
proto.removeUnderline = command( 'changeFormat', null, { tag: 'U' } );
proto.removeStrikethrough = command( 'changeFormat', null, { tag: 'S' } );
proto.removeSubscript = command( 'changeFormat', null, { tag: 'SUB' } );
proto.removeSuperscript = command( 'changeFormat', null, { tag: 'SUP' } );

proto.makeLink = function ( url, attributes ) {
    var range = this.getSelection();
    if ( range.collapsed ) {
        var protocolEnd = url.indexOf( ':' ) + 1;
        if ( protocolEnd ) {
            while ( url[ protocolEnd ] === '/' ) { protocolEnd += 1; }
        }
        insertNodeInRange(
            range,
            this._doc.createTextNode( url.slice( protocolEnd ) )
        );
    }
    attributes = mergeObjects(
        mergeObjects({
            href: url
        }, attributes, true ),
        this._config.tagAttributes.a,
        false
    );

    this.changeFormat({
        tag: 'A',
        attributes: attributes
    }, {
        tag: 'A'
    }, range );
    return this.focus();
};
proto.removeLink = function () {
    this.changeFormat( null, {
        tag: 'A'
    }, this.getSelection(), true );
    return this.focus();
};

proto.setFontFace = function ( name ) {
    this.changeFormat( name ? {
        tag: 'SPAN',
        attributes: {
            'class': FONT_FAMILY_CLASS,
            style: 'font-family: ' + name + ', sans-serif;'
        }
    } : null, {
        tag: 'SPAN',
        attributes: { 'class': FONT_FAMILY_CLASS }
    });
    return this.focus();
};
proto.setFontSize = function ( size ) {
    this.changeFormat( size ? {
        tag: 'SPAN',
        attributes: {
            'class': FONT_SIZE_CLASS,
            style: 'font-size: ' +
                ( typeof size === 'number' ? size + 'px' : size )
        }
    } : null, {
        tag: 'SPAN',
        attributes: { 'class': FONT_SIZE_CLASS }
    });
    return this.focus();
};

proto.setTextColour = function ( colour ) {
    this.changeFormat( colour ? {
        tag: 'SPAN',
        attributes: {
            'class': COLOUR_CLASS,
            style: 'color:' + colour
        }
    } : null, {
        tag: 'SPAN',
        attributes: { 'class': COLOUR_CLASS }
    });
    return this.focus();
};

proto.setHighlightColour = function ( colour ) {
    this.changeFormat( colour ? {
        tag: 'SPAN',
        attributes: {
            'class': HIGHLIGHT_CLASS,
            style: 'background-color:' + colour
        }
    } : colour, {
        tag: 'SPAN',
        attributes: { 'class': HIGHLIGHT_CLASS }
    });
    return this.focus();
};

proto.setTextAlignment = function ( alignment ) {
    this.forEachBlock( function ( block ) {
        var className = block.className
            .split( /\s+/ )
            .filter( function ( klass ) {
                return !!klass && !/^align/.test( klass );
            })
            .join( ' ' );
        if ( alignment ) {
            block.className = className + ' align-' + alignment;
            block.style.textAlign = alignment;
        } else {
            block.className = className;
            block.style.textAlign = '';
        }
    }, true );
    return this.focus();
};

proto.setTextDirection = function ( direction ) {
    this.forEachBlock( function ( block ) {
        if ( direction ) {
            block.dir = direction;
        } else {
            block.removeAttribute( 'dir' );
        }
    }, true );
    return this.focus();
};

function removeFormatting ( self, root, clean ) {
    var node, next;
    for ( node = root.firstChild; node; node = next ) {
        next = node.nextSibling;
        if ( isInline( node ) ) {
            if ( node.nodeType === TEXT_NODE || node.nodeName === 'BR' || node.nodeName === 'IMG' ) {
                clean.appendChild( node );
                continue;
            }
        } else if ( isBlock( node ) ) {
            clean.appendChild( self.createDefaultBlock([
                removeFormatting(
                    self, node, self._doc.createDocumentFragment() )
            ]));
            continue;
        }
        removeFormatting( self, node, clean );
    }
    return clean;
}

proto.removeAllFormatting = function ( range ) {
    if ( !range && !( range = this.getSelection() ) || range.collapsed ) {
        return this;
    }

    var root = this._root;
    var stopNode = range.commonAncestorContainer;
    while ( stopNode && !isBlock( stopNode ) ) {
        stopNode = stopNode.parentNode;
    }
    if ( !stopNode ) {
        expandRangeToBlockBoundaries( range, root );
        stopNode = root;
    }
    if ( stopNode.nodeType === TEXT_NODE ) {
        return this;
    }

    // Record undo point
    this.saveUndoState( range );

    // Avoid splitting where we're already at edges.
    moveRangeBoundariesUpTree( range, stopNode, stopNode, root );

    // Split the selection up to the block, or if whole selection in same
    // block, expand range boundaries to ends of block and split up to root.
    var doc = stopNode.ownerDocument;
    var startContainer = range.startContainer;
    var startOffset = range.startOffset;
    var endContainer = range.endContainer;
    var endOffset = range.endOffset;

    // Split end point first to avoid problems when end and start
    // in same container.
    var formattedNodes = doc.createDocumentFragment();
    var cleanNodes = doc.createDocumentFragment();
    var nodeAfterSplit = split( endContainer, endOffset, stopNode, root );
    var nodeInSplit = split( startContainer, startOffset, stopNode, root );
    var nextNode, childNodes;

    // Then replace contents in split with a cleaned version of the same:
    // blocks become default blocks, text and leaf nodes survive, everything
    // else is obliterated.
    while ( nodeInSplit !== nodeAfterSplit ) {
        nextNode = nodeInSplit.nextSibling;
        formattedNodes.appendChild( nodeInSplit );
        nodeInSplit = nextNode;
    }
    removeFormatting( this, formattedNodes, cleanNodes );
    cleanNodes.normalize();
    nodeInSplit = cleanNodes.firstChild;
    nextNode = cleanNodes.lastChild;

    // Restore selection
    childNodes = stopNode.childNodes;
    if ( nodeInSplit ) {
        stopNode.insertBefore( cleanNodes, nodeAfterSplit );
        startOffset = indexOf.call( childNodes, nodeInSplit );
        endOffset = indexOf.call( childNodes, nextNode ) + 1;
    } else {
        startOffset = indexOf.call( childNodes, nodeAfterSplit );
        endOffset = startOffset;
    }

    // Merge text nodes at edges, if possible
    range.setStart( stopNode, startOffset );
    range.setEnd( stopNode, endOffset );
    mergeInlines( stopNode, range );

    // And move back down the tree
    moveRangeBoundariesDownTree( range );

    this.setSelection( range );
    this._updatePath( range, true );

    return this.focus();
};

proto.increaseQuoteLevel = command( 'modifyBlocks', increaseBlockQuoteLevel );
proto.decreaseQuoteLevel = command( 'modifyBlocks', decreaseBlockQuoteLevel );

proto.makeUnorderedList = command( 'modifyBlocks', makeUnorderedList );
proto.makeOrderedList = command( 'modifyBlocks', makeOrderedList );
proto.removeList = command( 'modifyBlocks', removeList );

// Node.js exports
Squire.isInline = isInline;
Squire.isBlock = isBlock;
Squire.isContainer = isContainer;
Squire.getBlockWalker = getBlockWalker;
Squire.getPreviousBlock = getPreviousBlock;
Squire.getNextBlock = getNextBlock;
Squire.areAlike = areAlike;
Squire.hasTagAttributes = hasTagAttributes;
Squire.getNearest = getNearest;
Squire.isOrContains = isOrContains;
Squire.detach = detach;
Squire.replaceWith = replaceWith;
Squire.empty = empty;

// Range.js exports
Squire.getNodeBefore = getNodeBefore;
Squire.getNodeAfter = getNodeAfter;
Squire.insertNodeInRange = insertNodeInRange;
Squire.extractContentsOfRange = extractContentsOfRange;
Squire.deleteContentsOfRange = deleteContentsOfRange;
Squire.insertTreeFragmentIntoRange = insertTreeFragmentIntoRange;
Squire.isNodeContainedInRange = isNodeContainedInRange;
Squire.moveRangeBoundariesDownTree = moveRangeBoundariesDownTree;
Squire.moveRangeBoundariesUpTree = moveRangeBoundariesUpTree;
Squire.getStartBlockOfRange = getStartBlockOfRange;
Squire.getEndBlockOfRange = getEndBlockOfRange;
Squire.contentWalker = contentWalker;
Squire.rangeDoesStartAtBlockBoundary = rangeDoesStartAtBlockBoundary;
Squire.rangeDoesEndAtBlockBoundary = rangeDoesEndAtBlockBoundary;
Squire.expandRangeToBlockBoundaries = expandRangeToBlockBoundaries;

// Clipboard.js exports
Squire.onPaste = onPaste;

// Editor.js exports
Squire.addLinks = addLinks;
Squire.splitBlock = splitBlock;
Squire.startSelectionId = startSelectionId;
Squire.endSelectionId = endSelectionId;

if ( typeof exports === 'object' ) {
    module.exports = Squire;
} else if ( typeof define === 'function' && define.amd ) {
    define( function () {
        return Squire;
    });
} else {
    win.Squire = Squire;

    if ( top !== win &&
            doc.documentElement.getAttribute( 'data-squireinit' ) === 'true' ) {
        win.editor = new Squire( doc );
        if ( win.onEditorLoad ) {
            win.onEditorLoad( win.editor );
            win.onEditorLoad = null;
        }
    }
}

}( document ) );
