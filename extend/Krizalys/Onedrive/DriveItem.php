<?php

namespace Krizalys\Onedrive;

/**
 * @class DriveItem
 *
 * A DriveItem instance is an entity that may be stored in a OneDrive account.
 * There are two types of drive items: file or a folder, each of which being a
 * subclass of the DriveItem class.
 *
 * Note that DriveItem instances are only "proxy" to actual OneDrive drive items
 * (eg. destroying a DriveItem instance will not delete the actual OneDrive
 * drive item it is referencing to).
 */
abstract class DriveItem
{
    /**
     * @var Client The owning Client instance.
     */
    protected $_client;

    /**
     * @var string The unique ID assigned by OneDrive to this drive item.
     */
    protected $_id;

    /**
     * @var string The unique ID assigned by OneDrive to the parent folder of
     *             this drive item.
     */
    private $_parentId;

    /**
     * @var string The name of this drive item.
     */
    private $_name;

    /**
     * @var string The description of this drive item.
     */
    private $_description;

    /**
     * @var int The size of this drive item, in bytes.
     */
    private $_size;

    /**
     * @var string The source link of this drive item.
     */
    private $_source;

    /**
     * @var int The creation time, in seconds since UNIX epoch.
     */
    private $_createdTime;

    /**
     * @var int The last modification time, in seconds since UNIX epoch.
     */
    private $_updatedTime;

    /**
     * Constructor.
     *
     * @param Client       $client  The Client instance owning this DriveItem
     *                              instance.
     * @param null|string  $id      The unique ID of the OneDrive drive item
     *                              referenced by this DriveItem instance.
     * @param array|object $options An array/object with one or more of the
     *                              following keys/properties:
     *                              - 'parent_id'    (string) The unique ID of
     *                              the parent OneDrive folder of this drive
     *                              item.
     *                              - 'name' (string) The name of this drive
     *                              item.
     *                              - 'description' (string) The description of
     *                              this drive item. May be empty.
     *                              - 'size' (int) The size of this drive item,
     *                              in bytes.
     *                              - 'source' (string) The source link of this
     *                              drive item.
     *                              - 'created_time' (string) The creation time,
     *                              as a RFC date/time.
     *                              - 'updated_time' (string) The last
     *                              modification time, as a RFC date/time.
     */
    public function __construct(Client $client, $id, $options = [])
    {
        $options       = (object) $options;
        $this->_client = $client;
        $this->_id     = null !== $id ? (string) $id : null;

        $this->_parentId = property_exists($options, 'parent_id') ?
            (string) $options->parent_id : null;

        $this->_name = property_exists($options, 'name') ?
            (string) $options->name : null;

        $this->_description = property_exists($options, 'description') ?
            (string) $options->description : null;

        $this->_size = property_exists($options, 'size') ?
            (int) $options->size : null;

        $this->_source = property_exists($options, 'source') ?
            (string) $options->source : null;

        $this->_createdTime = property_exists($options, 'created_time') ?
            strtotime($options->created_time) : null;

        $this->_updatedTime = property_exists($options, 'updated_time') ?
            strtotime($options->updated_time) : null;
    }

    /**
     * Determines whether the OneDrive drive item referenced by this DriveItem
     * instance is a folder.
     *
     * @return bool true if the OneDrive drive item referenced by this DriveItem
     *              instance is a folder, false otherwise.
     */
    public function isFolder()
    {
        return false;
    }

    /**
     * Fetches the properties of the OneDrive drive item referenced by this
     * DriveItem instance. Some properties are cached for faster subsequent
     * access.
     *
     * @return array The properties of the OneDrive drive item referenced by
     *               this DriveItem instance.
     */
    public function fetchProperties()
    {
        $result = $this->_client->fetchProperties($this->_id);

        $this->_parentId = '' != $result->parent_id ?
            (string) $result->parent_id : null;

        $this->_name = $result->name;

        $this->_description = '' != $result->description ?
            (string) $result->description : null;

        $this->_size = (int) $result->size;

        /** @todo Handle volatile existence (eg. present only for files). */
        $this->_source = (string) $result->source;

        $this->_createdTime = strtotime($result->created_time);
        $this->_updatedTime = strtotime($result->updated_time);

        return $result;
    }

    /**
     * Gets the unique ID of the OneDrive drive item referenced by this
     * DriveItem instance.
     *
     * @return string The unique ID of the OneDrive drive item referenced by
     *                this DriveItem instance.
     */
    public function getId()
    {
        return $this->_id;
    }

    /**
     * Gets the unique ID of the parent folder of the OneDrive drive item
     * referenced by this DriveItem instance.
     *
     * @return string The unique ID of the OneDrive folder containing the drive
     *                item referenced by this DriveItem instance.
     */
    public function getParentId()
    {
        if (null === $this->_parentId) {
            $this->fetchProperties();
        }

        return $this->_parentId;
    }

    /**
     * Gets the name of the OneDrive drive item referenced by this DriveItem
     * instance.
     *
     * @return string The name of the OneDrive drive item referenced by this
     *                DriveItem instance.
     */
    public function getName()
    {
        if (null === $this->_name) {
            $this->fetchProperties();
        }

        return $this->_name;
    }

    /**
     * Gets the description of the OneDrive drive item referenced by this
     * DriveItem instance.
     *
     * @return string The description of the OneDrive drive item referenced by
     *                this DriveItem instance.
     */
    public function getDescription()
    {
        if (null === $this->_description) {
            $this->fetchProperties();
        }

        return $this->_description;
    }

    /**
     * Gets the size of the OneDrive drive item referenced by this DriveItem
     * instance.
     *
     * @return int The size of the OneDrive drive item referenced by this
     *             DriveItem instance.
     */
    public function getSize()
    {
        if (null === $this->_size) {
            $this->fetchProperties();
        }

        return $this->_size;
    }

    /**
     * Gets the source link of the OneDrive drive item referenced by this
     * DriveItem instance.
     *
     * @return string The source link of the OneDrive drive item referenced by
     *                this DriveItem instance.
     */
    public function getSource()
    {
        if (null === $this->_source) {
            $this->fetchProperties();
        }

        return $this->_source;
    }

    /**
     * Gets the creation time of the OneDrive drive item referenced by this
     * DriveItem instance.
     *
     * @return int The creation time of the drive item referenced by this
     *             DriveItem instance, in seconds since UNIX epoch.
     */
    public function getCreatedTime()
    {
        if (null === $this->_createdTime) {
            $this->fetchProperties();
        }

        return $this->_createdTime;
    }

    /**
     * Gets the last modification time of the OneDrive drive item referenced by
     * this DriveItem instance.
     *
     * @return int The last modification time of the drive item referenced by
     *             this DriveItem instance, in seconds since UNIX epoch.
     */
    public function getUpdatedTime()
    {
        if (null === $this->_updatedTime) {
            $this->fetchProperties();
        }

        return $this->_updatedTime;
    }

    /**
     * Moves the OneDrive drive item referenced by this DriveItem instance into
     * another OneDrive folder.
     *
     * @param null|string $destinationId The unique ID of the OneDrive folder
     *                                   into which to move the OneDrive drive
     *                                   item referenced by this DriveItem
     *                                   instance, or null to move it to the
     *                                   OneDrive root folder. Default: null.
     */
    public function move($destinationId = null)
    {
        $this->_client->moveObject($this->_id, $destinationId);
    }
}
