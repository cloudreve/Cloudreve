<?php

namespace Krizalys\Onedrive;

/**
 * @class Folder
 *
 * A Folder instance is a DriveItem instance referencing to a OneDrive folder.
 * It may contain other OneDrive drive items but may not have content.
 */
class Folder extends DriveItem
{
    /**
     * Determines whether the OneDrive drive item referenced by this DriveItem
     * instance is a folder.
     *
     * @return bool true if the OneDrive drive item referenced by this DriveItem
     *              instance is a folder, false otherwise.
     */
    public function isFolder()
    {
        return true;
    }

    /**
     * Constructor.
     *
     * @param Client       $client  The Client instance owning this DriveItem
     *                              instance.
     * @param null|string  $id      The unique ID of the OneDrive drive item
     *                              referenced by this DriveItem instance, or
     *                              null to reference the OneDrive root folder.
     *                              Default: null.
     * @param array|object $options Options to pass to the DriveItem
     *                              constructor.
     */
    public function __construct(Client $client, $id = null, $options = [])
    {
        parent::__construct($client, $id, $options);
    }

    /**
     * Gets the drive items in the OneDrive folder referenced by this Folder
     * instance.
     *
     * @return array The drive items in the OneDrive folder referenced by this
     *               Folder instance, as DriveItem instances.
     *
     * @deprecated Use Folder::fetchChildObjects() instead.
     */
    public function fetchObjects()
    {
        /** @todo Log deprecation notice. */
        return $this->fetchChildObjects();
    }

    /**
     * Gets the child drive items in the OneDrive folder referenced by this
     * Folder instance.
     *
     * @return array The drive items in the OneDrive folder referenced by this
     *               Folder instance, as DriveItem instances.
     */
    public function fetchChildObjects()
    {
        return $this->_client->fetchObjects($this->_id);
    }

    /**
     * Gets the descendant drive items under the OneDrive folder referenced by
     * this Folder instance.
     *
     * @return array The files in the OneDrive folder referenced by this Folder
     *               instance, as DriveItem instances.
     */
    public function fetchDescendantObjects()
    {
        $driveItems = [];

        foreach ($this->fetchChildObjects() as $driveItem) {
            if ($driveItem->isFolder()) {
                $driveItems = array_merge(
                    $driveItem->fetchDescendantObjects(),
                    $driveItems
                );
            } else {
                array_push($driveItems, $driveItem);
            }
        }

        return $driveItems;
    }

    /**
     * Creates a folder in the OneDrive folder referenced by this Folder
     * instance.
     *
     * @param string      $name        The name of the OneDrive folder to be
     *                                 created.
     * @param null|string $description The description of the OneDrive folder
     *                                 to be created, or null to create it
     *                                 without a description. Default: null.
     *
     * @return Folder The folder created, as a Folder instance.
     */
    public function createFolder($name, $description = null)
    {
        return $this->_client->createFolder($name, $this->_id, $description);
    }

    /**
     * Creates a file in the OneDrive folder referenced by this Folder instance.
     *
     * @param string          $name    The name of the OneDrive file to be
     *                                 created.
     * @param string|resource $content The content of the OneDrive file to be
     *                                 created, as a string or handle to an
     *                                 already opened file.  In the latter case,
     *                                 the responsibility to close the handle is
     *                                 is left to the calling function. Default:
     *                                 ''.
     * @param array           $options The options.
     *
     * @return File The file created, as a File instance.
     *
     * @throws \Exception Thrown on I/O errors.
     */
    public function createFile($name, $content = '', array $options = [])
    {
        $client = $this->_client;

        $options = array_merge([
            'name_conflict_behavior' => $client->getNameConflictBehavior(),
            'stream_back_end'        => $client->getStreamBackEnd(),
        ], $options);

        return $this->_client->createFile(
            $name, $this->_id,
            $content,
            $options
        );
    }
}
