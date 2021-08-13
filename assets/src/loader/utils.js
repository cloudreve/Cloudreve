export const isDefined = (val) => val != null;
export const isFunction = (val) => typeof val === "function";
// eslint-disable-next-line @typescript-eslint/no-empty-function
export const noop = () => {};

export const newScript = (src) => (cb) => {
    const scriptElem = document.createElement("script");
    if (typeof src === "object") {
        // copy every property to the element
        for (const key in src) {
            if (Object.prototype.hasOwnProperty.call(src, key)) {
                scriptElem[key] = src[key];
            }
        }
        src = src.src;
    } else {
        scriptElem.src = src;
    }
    scriptElem.addEventListener("load", () => cb(null, src));
    scriptElem.addEventListener("error", () => cb(true, src));
    document.body.appendChild(scriptElem);
    return scriptElem;
};

const keyIterator = (cols) => {
    const keys = Object.keys(cols);
    let i = -1;
    return {
        next() {
            i++; // inc
            if (i >= keys.length) return null;
            else return keys[i];
        },
    };
};

// tasks should be a collection of thunk
export const parallel = (...tasks) => (each) => (cb) => {
    let hasError = false;
    let successed = 0;
    const ret = [];
    tasks = tasks.filter(isFunction);

    if (tasks.length <= 0) cb(null);
    else {
        tasks.forEach((task, i) => {
            const thunk = task;
            thunk((err, ...args) => {
                if (err) hasError = true;
                else {
                    // collect result
                    if (args.length <= 1) args = args[0];

                    ret[i] = args;
                    successed++;
                }

                if (isFunction(each)) each.call(null, err, args, i);

                if (hasError) cb(true);
                else if (tasks.length === successed) {
                    cb(null, ret);
                }
            });
        });
    }
};

// tasks should be a collection of thunk
export const series = (...tasks) => (each) => (cb) => {
    tasks = tasks.filter((val) => val != null);
    const nextKey = keyIterator(tasks);
    const nextThunk = () => {
        const key = nextKey.next();
        let thunk = tasks[key];
        // eslint-disable-next-line prefer-spread
        if (Array.isArray(thunk))
            // eslint-disable-next-line prefer-spread
            thunk = parallel.apply(null, thunk).call(null, each);
        return [+key, thunk]; // convert `key` to number
    };
    let key, thunk;
    let next = nextThunk();
    key = next[0];
    thunk = next[1];
    if (thunk == null) return cb(null);

    const ret = [];
    const iterator = () => {
        thunk((err, ...args) => {
            if (args.length <= 1) args = args[0];
            if (isFunction(each)) each.call(null, err, args, key);

            if (err) cb(err);
            else {
                // collect result
                ret.push(args);

                next = nextThunk();
                key = next[0];
                thunk = next[1];
                if (thunk == null) return cb(null, ret);
                // finished
                else iterator();
            }
        });
    };
    iterator();
};
