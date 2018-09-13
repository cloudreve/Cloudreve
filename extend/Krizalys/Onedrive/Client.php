<?php
namespace Krizalys\Onedrive;
/**
 * @class Client
 *
 * A Client instance allows communication with the OneDrive API and perform
 * operations programmatically.
 *
 * For an overview of the OneDrive protocol flow, see here:
 * http://msdn.microsoft.com/en-us/library/live/hh243647.aspx
 *
 * To manage your Live Connect applications, see here:
 * https://account.live.com/developers/applications/index
 * Or here:
 * https://manage.dev.live.com/ (not working?)
 *
 * For an example implementation, see here:
 * https://github.com/drumaddict/skydrive-api-yii/blob/master/SkyDriveAPI.php
 */
class Client
{
    /**
     * @var string The base URL for API requests.
     */
    const API_URL = 'https://graph.microsoft.com/v1.0';

    /**
     * @var string The base URL for authorization requests.
     */
    const AUTH_URL = 'https://login.microsoftonline.com/common/oauth2/v2.0';

    /**
     * @var string The base URL for token requests.
     */
    const TOKEN_URL = 'https://login.live.com/oauth20_token.srf';

    /**
     * @var string Client information.
     */
    private $_clientId;

    /**
     * @var object OAuth state (token, etc...).
     */
    private $_state;

    /**
     * @var int The last HTTP status received.
     */
    private $_httpStatus;

    /**
     * @var string The last Content-Type received.
     */
    private $_contentType;

    /**
     * @var bool Verify SSL hosts and peers.
     */
    private $_sslVerify;

    /**
     * @var null|string Over-ride SSL CA path for verification (only relevant
     *                  when verifying).
     */
    private $_sslCaPath;

    /**
     * @var int The name conflict behavior.
     */
    private $_nameConflictBehavior;

    /**
     * @var int The stream back end.
     */
    private $_streamBackEnd;

    /**
     * @var NameConflictBehaviorParameterizer The name conflict behavior
     *                                        parameterizer.
     */
    private $_nameConflictBehaviorParameterizer;

    /**
     * @var StreamOpener The stream opener.
     */
    private $_streamOpener;

    /**
     * Creates a base cURL object which is compatible with the OneDrive API.
     *
     * @param string $path    The path of the API call (eg. me/skydrive).
     * @param array  $options Extra cURL options to apply.
     *
     * @return resource A compatible cURL object.
     */
    private function _createCurl($path, $options = [])
    {
        $curl = curl_init();

        $defaultOptions = [
            // General options.
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_AUTOREFERER    => true,

            // SSL options.
            // The value 2 checks the existence of a common name and also
            // verifies that it matches the hostname provided.
            CURLOPT_SSL_VERIFYHOST => ($this->_sslVerify ? 2 : false),

            CURLOPT_SSL_VERIFYPEER => $this->_sslVerify,
        ];

        if ($this->_sslVerify && $this->_sslCaPath) {
            $defaultOptions[CURLOPT_CAINFO] = $this->_sslCaPath;
        }

        // See http://php.net/manual/en/function.array-merge.php for a
        // description of the + operator (and why array_merge() would be wrong).
        $finalOptions = $options + $defaultOptions;

        curl_setopt_array($curl, $finalOptions);
        return $curl;
    }

    /**
     * Processes a result returned by the OneDrive API call using a cURL object.
     *
     * @param resource $curl The cURL object used to perform the call.
     *
     * @return object|string The content returned, as an object instance if
     *                       served a JSON, or as a string if served as anything
     *                       else.
     *
     * @throws \Exception Thrown if curl_exec() fails.
     */
    private function _processResult($curl)
    {
        $result = curl_exec($curl);

        if (false === $result) {
            throw new \Exception('curl_exec() failed: ' . curl_error($curl));
        }

        $info = curl_getinfo($curl);

        $this->_httpStatus = array_key_exists('http_code', $info) ?
            (int) $info['http_code'] : null;

        $this->_contentType = array_key_exists('content_type', $info) ?
            (string) $info['content_type'] : null;

        // Parse nothing but JSON.
        if (1 !== preg_match('|^application/json|', $this->_contentType)) {
            return $result;
        }

        // Empty JSON string is returned as an empty object.
        if ('' == $result) {
            return (object) [];
        }

        $decoded = json_decode($result);
        $vars    = get_object_vars($decoded);

        if (array_key_exists('error', $vars)) {
            throw new \Exception($decoded->error->message,
                (int) $decoded->error->code);
        }

        return $decoded;
    }

    /**
     * Constructor.
     *
     * @param array $options The options to use while creating this object.
     *                       Valid supported keys are:
     *                       - 'state' (object) When defined, it should contain
     *                       a valid OneDrive client state, as returned by
     *                       getState(). Default: [].
     *                       - 'ssl_verify' (bool) Whether to verify SSL hosts
     *                       and peers. Default: false.
     *                       - 'ssl_capath' (bool|string) CA path to use for
     *                       verifying SSL certificate chain. Default: false.
     *                       - 'name_conflict_behavior' (int) Default name
     *                       conflict behavior. Either:
     *                       NameConflictBehavior::FAIL,
     *                       NameConflictBehavior::RENAME or
     *                       NameConflictBehavior::REPLACE. Default:
     *                       NameConflictBehavior::REPLACE.
     *                       - 'stream_back_end' (int) Default stream back end.
     *                       Either StreamBackEnd::MEMORY or
     *                       StreamBackEnd::TEMP. Default:
     *                       StreamBackEnd::MEMORY.
     *                       Using temporary files is recommended when uploading
     *                       big files.
     *                       Default: StreamBackEnd::MEMORY.
     */
    public function __construct(array $options = [])
    {
        $this->_clientId = array_key_exists('client_id', $options)
            ? (string) $options['client_id'] : null;

        $this->_state = array_key_exists('state', $options)
            ? $options['state'] : (object) [
                'redirect_uri' => null,
                'token'        => null,
            ];

        $this->_sslVerify = array_key_exists('ssl_verify', $options)
            ? $options['ssl_verify'] : false;

        $this->_sslCaPath = array_key_exists('ssl_capath', $options)
            ? $options['ssl_capath'] : false;

        $this->_nameConflictBehavior =
            array_key_exists('name_conflict_behavior', $options) ?
            $options['name_conflict_behavior']
            : NameConflictBehavior::REPLACE;

        $this->_streamBackEnd = array_key_exists('stream_back_end', $options)
            ? $options['stream_back_end'] : StreamBackEnd::MEMORY;

        $this->_nameConflictBehaviorParameterizer =
            new NameConflictBehaviorParameterizer();

        $this->_streamOpener = new StreamOpener();
    }

    /**
     * Gets the name conflict behavior of this client instance.
     *
     * @return int
     */
    public function getNameConflictBehavior()
    {
        return $this->_nameConflictBehavior;
    }

    /**
     * Gets the stream back end of this client instance.
     *
     * @return int
     */
    public function getStreamBackEnd()
    {
        return $this->_streamBackEnd;
    }

    /**
     * Gets the current state of this Client instance. Typically saved in the
     * session and passed back to the Client constructor for further requests.
     *
     * @return object The state of this Client instance.
     */
    public function getState()
    {
        return $this->_state;
    }

    /**
     * Gets the URL of the log in form. After login, the browser is redirected
     * to the redirect URL, and a code is passed as a GET parameter to this URL.
     *
     * The browser is also redirected to this URL if the user is already logged
     * in.
     *
     * @param array  $scopes      The OneDrive scopes requested by the
     *                            application. Supported values:
     *                            - 'wl.signin'
     *                            - 'wl.basic'
     *                            - 'wl.contacts_skydrive'
     *                            - 'wl.skydrive_update'
     * @param string $redirectUri The URI to which to redirect to upon
     *                            successful log in.
     * @param array  $options     Reserved for future use. Default: [].
     *
     * @return string The login URL.
     *
     * @throws \Exception Thrown if this Client instance's clientId is not set.
     *
     * @todo Support $options.
     */
    public function getLogInUrl(
        array $scopes,
        $redirectUri,
        array $options = []
    ) {
        if (null === $this->_clientId) {
            throw new \Exception(
                'The client ID must be set to call getLoginUrl()'
            );
        }

        $imploded                   = implode('+', $scopes);
        $redirectUri                = (string) $redirectUri;
        $this->_state->redirect_uri = $redirectUri;

        // When using this URL, the browser will eventually be redirected to the
        // callback URL with a code passed in the URL query string (the name of
        // the variable is "code"). This is suitable for PHP.
        $url = self::AUTH_URL
            . '/authorize?client_id=' . urlencode($this->_clientId)
            . '&scope=' . $imploded
            . '&response_type=code'
            . '&redirect_uri=' . urlencode($redirectUri)
            . '&display=popup'
            . '&locale=en';

        return $url;
    }

    /**
     * Gets the access token expiration delay.
     *
     * @return int The token expiration delay, in seconds.
     */
    public function getTokenExpire()
    {
        return $this->_state->token->obtained
            + $this->_state->token->data->expires_in - time();
    }

    /**
     * Gets the status of the current access token.
     *
     * @return int The status of the current access token:
     *             -  0 No access token.
     *             - -1 Access token will expire soon (1 minute or less).
     *             - -2 Access token is expired.
     *             -  1 Access token is valid.
     */
    public function getAccessTokenStatus()
    {
        if (null === $this->_state->token) {
            return 0;
        }

        $remaining = $this->getTokenExpire();

        if (0 >= $remaining) {
            return -2;
        }

        if (60 >= $remaining) {
            return -1;
        }

        return 1;
    }

    /**
     * Obtains a new access token from OAuth. This token is valid for one hour.
     *
     * @param string $clientSecret The OneDrive client secret.
     * @param string $code         The code returned by OneDrive after
     *                             successful log in.
     * @param string $redirectUri  Must be the same as the redirect URI passed
     *                             to getLoginUrl().
     *
     * @throws \Exception Thrown if this Client instance's clientId is not set.
     * @throws \Exception Thrown if the redirect URI of this Client instance's
     *                    state is not set.
     */
    public function obtainAccessToken($clientSecret, $code)
    {
        if (null === $this->_clientId) {
            throw new \Exception(
                'The client ID must be set to call obtainAccessToken()'
            );
        }

        if (null === $this->_state->redirect_uri) {
            throw new \Exception(
                'The state\'s redirect URI must be set to call'
                    . ' obtainAccessToken()'
            );
        }

        $url = self::AUTH_URL."/token";

        $curl = curl_init();

        $fields = http_build_query(
            [
                'client_id'     => $this->_clientId,
                'redirect_uri'  => $this->_state->redirect_uri,
                'client_secret' => $clientSecret,
                'code'          => $code,
                'grant_type'    => 'authorization_code',
            ]
        );

        curl_setopt_array($curl, [
            // General options.
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_AUTOREFERER    => true,
            CURLOPT_POST           => 1,
            CURLOPT_POSTFIELDS     => $fields,

            CURLOPT_HTTPHEADER => [
                'Content-Length: ' . strlen($fields),
            ],

            // SSL options.
            CURLOPT_SSL_VERIFYHOST => false,
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_URL            => $url,
        ]);

        $result = curl_exec($curl);

        if (false === $result) {
            if (curl_errno($curl)) {
                throw new \Exception('curl_setopt_array() failed: '
                    . curl_error($curl));
            } else {
                throw new \Exception('curl_setopt_array(): empty response');
            }
        }

        $decoded = json_decode($result);

        if (null === $decoded) {
            throw new \Exception('json_decode() failed');
        }

        $this->_state->redirect_uri = null;

        $this->_state->token = (object) [
            'obtained' => time(),
            'data'     => $decoded,
        ];
    }

    /**
     * Renews the access token from OAuth. This token is valid for one hour.
     *
     * @param string $clientSecret The client secret.
     */
    public function renewAccessToken($clientSecret)
    {
        if (null === $this->_clientId) {
            throw new \Exception(
                'The client ID must be set to call renewAccessToken()'
            );
        }

        if (null === $this->_state->token->data->refresh_token) {
            throw new \Exception(
                'The refresh token is not set or no permission for'
                    . ' \'wl.offline_access\' was given to renew the token'
            );
        }

        $url = self::AUTH_URL."/token";

        $curl = curl_init();

        curl_setopt_array($curl, [
            // General options.
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_FOLLOWLOCATION => true,
            CURLOPT_AUTOREFERER    => true,
            CURLOPT_POST           => 1, // i am sending post data

            CURLOPT_POSTFIELDS =>
                'client_id=' . urlencode($this->_clientId)
                    . '&client_secret=' . urlencode($clientSecret)
                    . '&grant_type=refresh_token'
                    . '&refresh_token='.$this->_state->token->data->refresh_token,

            // SSL options.
            CURLOPT_SSL_VERIFYHOST => false,
            CURLOPT_SSL_VERIFYPEER => false,
            CURLOPT_URL            => $url,
        ]);

        $result = curl_exec($curl);

        if (false === $result) {
            if (curl_errno($curl)) {
                throw new \Exception(
                    'curl_setopt_array() failed: ' . curl_error($curl)
                );
            } else {
                throw new \Exception('curl_setopt_array(): empty response');
            }
        }

        $decoded = json_decode($result);

        if (null === $decoded) {
            throw new \Exception('json_decode() failed');
        }

        $this->_state->token = (object) [
            'obtained' => time(),
            'data'     => $decoded,
        ];
    }

    /**
     * Performs a call to the OneDrive API using the GET method.
     *
     * @param string $path    The path of the API call (eg. me/skydrive).
     * @param array  $options Further curl options to set.
     *
     * @return object|string The response body, if any.
     */
    public function apiGet($path, $options = [])
    {
        $url =
            self::API_URL
                . $path;
        $curl = self::_createCurl($path, $options);
        curl_setopt($curl, CURLOPT_HTTPHEADER, [ 'Authorization: Bearer '
        . $this->_state->token->data->access_token,]);
        curl_setopt($curl, CURLOPT_URL, $url);
        return $this->_processResult($curl);
    }

    /**
     * Performs a call to the OneDrive API using the POST method.
     *
     * @param string       $path The path of the API call (eg. me/skydrive).
     * @param array|object $data The data to pass in the body of the request.
     *
     * @return object|string The response body, if any.
     */
    public function apiPost($path, $data)
    {
        $url  = self::API_URL . $path;
        $data = (object) $data;
        $curl = self::_createCurl($path);
        curl_setopt_array($curl, [
            CURLOPT_URL        => $url,
            CURLOPT_POST       => true,

            CURLOPT_HTTPHEADER => [
                // The data is sent as JSON as per OneDrive documentation.
                'Content-Type: application/json',

                'Authorization: Bearer '
                    . $this->_state->token->data->access_token,
            ],

            CURLOPT_POSTFIELDS => json_encode($data),
        ]);

        return $this->_processResult($curl);
    }

    /**
     * Performs a call to the OneDrive API using the PUT method.
     *
     * @param string   $path        The path of the API call (eg. me/skydrive).
     * @param resource $stream      The data stream to upload.
     * @param string   $contentType The MIME type of the data stream, or null if
     *                              unknown. Default: null.
     *
     * @return object|string The response body, if any.
     */
    public function apiPut($path, $stream, $contentType = null)
    {
        $url   = self::API_URL . $path;
        $curl  = self::_createCurl($path);
        $stats = fstat($stream);

        $headers = [
            'Authorization: Bearer ' . $this->_state->token->data->access_token,
        ];

        if (null !== $contentType) {
            $headers[] = 'Content-Type: ' . $contentType;
        }

        $options = [
            CURLOPT_URL        => $url,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_PUT        => true,
            CURLOPT_INFILE     => $stream,
            CURLOPT_INFILESIZE => $stats[7], // Size
        ];

        curl_setopt_array($curl, $options);
        return $this->_processResult($curl);
    }

    public function sendFileChunk($url,$headers,$stream){
        $curl  = self::_createCurl("");
        $close = 0;
        if(!is_resource($stream)){
            $close = 1;
            $options=[];
            $options = array_merge([
                'stream_back_end' => $this->_streamBackEnd,
            ], $options);
            $content = $stream;
            $stream = $this
                ->_streamOpener
                ->open($options['stream_back_end']);

            if (false === $stream) {
                throw new \Exception('fopen() failed');
            }

            if (false === fwrite($stream, $content)) {
                fclose($stream);
                throw new \Exception('fwrite() failed');
            }

            if (!rewind($stream)) {
                fclose($stream);
                throw new \Exception('rewind() failed');
            }
        }
        $stats = fstat($stream);

        $options = [
            CURLOPT_URL        => $url,
            CURLOPT_HTTPHEADER => $headers,
            CURLOPT_PUT        => true,
            CURLOPT_INFILE     => $stream,
            CURLOPT_INFILESIZE => $stats[7], // Size
        ];

        curl_setopt_array($curl, $options);
        
        $data = $this->_processResult($curl);
        if (is_resource($close)) {
            fclose($stream);
        }
        return $data;
    }

    /**
     * Performs a call to the OneDrive API using the DELETE method.
     *
     * @param string $path The path of the API call (eg. me/skydrive).
     *
     * @return object|string The response body, if any.
     */
    public function apiDelete($path)
    {
        $url =
            self::API_URL
                . $path; 
        $curl = self::_createCurl($path);

        curl_setopt_array($curl, [
            CURLOPT_URL           => $url,
            CURLOPT_CUSTOMREQUEST => 'DELETE',
            CURLOPT_HTTPHEADER =>[
                "Authorization: bearer ".$this->_state->token->data->access_token,
            ]

        ]);

        return $this->_processResult($curl);
    }

    /**
     * Performs a call to the OneDrive API using the MOVE method.
     *
     * @param string       $path The path of the API call (eg. me/skydrive).
     * @param array|object $data The data to pass in the body of the request.
     *
     * @return object|string The response body, if any.
     */
    public function apiMove($path, $data)
    {
        $url  = self::API_URL . $path;
        $data = (object) $data;
        $curl = self::_createCurl($path);

        curl_setopt_array($curl, [
            CURLOPT_URL           => $url,
            CURLOPT_CUSTOMREQUEST => 'MOVE',

            CURLOPT_HTTPHEADER    => [
                // The data is sent as JSON as per OneDrive documentation.
                'Content-Type: application/json',

                'Authorization: Bearer '
                    . $this->_state->token->data->access_token,
            ],

            CURLOPT_POSTFIELDS    => json_encode($data),
        ]);

        return $this->_processResult($curl);
    }

    /**
     * Performs a call to the OneDrive API using the COPY method.
     *
     * @param string       $path The path of the API call (eg. me/skydrive).
     * @param array|object $data The data to pass in the body of the request.
     *
     * @return object|string The response body, if any.
     */
    public function apiCopy($path, $data)
    {
        $url  = self::API_URL . $path;
        $data = (object) $data;
        $curl = self::_createCurl($path);

        curl_setopt_array($curl, [
            CURLOPT_URL           => $url,
            CURLOPT_CUSTOMREQUEST => 'COPY',

            CURLOPT_HTTPHEADER    => [
                // The data is sent as JSON as per OneDrive documentation.
                'Content-Type: application/json',

                'Authorization: Bearer '
                    . $this->_state->token->data->access_token,
            ],

            CURLOPT_POSTFIELDS    => json_encode($data),
        ]);

        return $this->_processResult($curl);
    }

    /**
     * Creates a folder in the current OneDrive account.
     *
     * @param string      $name        The name of the OneDrive folder to be
     *                                 created.
     * @param null|string $parentId    The ID of the OneDrive folder into which
     *                                 to create the OneDrive folder, or null to
     *                                 create it in the OneDrive root folder.
     *                                 Default: null.
     * @param null|string $description The description of the OneDrive folder to
     *                                 be created, or null to create it without
     *                                 a description. Default: null.
     *
     * @return Folder The folder created, as a Folder instance referencing to
     *                the OneDrive folder created.
     */
    public function createFolder($name, $parentId = null, $description = null)
    {
        if (null === $parentId) {
            $parentId = 'me/skydrive';
        }

        $properties = [
            'name' => (string) $name,
        ];

        if (null !== $description) {
            $properties['description'] = (string) $description;
        }

        $folder = $this->apiPost($parentId, (object) $properties);

        return new Folder($this, $folder->id, $folder);
    }

    /**
     * Creates a file in the current OneDrive account.
     *
     * @param string          $name     The name of the OneDrive file to be
     *                                  created.
     * @param null|string     $parentId The ID of the OneDrive folder into which
     *                                  to create the OneDrive file, or null to
     *                                  create it in the OneDrive root folder.
     *                                  Default: null.
     * @param string|resource $content  The content of the OneDrive file to be
     *                                  created, as a string or as a resource to
     *                                  an already opened file. In the latter
     *                                  case, the responsibility to close the
     *                                  handle is left to the calling function.
     *                                  Default: ''.
     * @param array           $options  The options.
     *
     * @return File The file created, as File instance referencing to the
     *              OneDrive file created.
     *
     * @throws \Exception Thrown on I/O errors.
     */
    public function createFile(
        $name,
        $parentId = null,
        $content = '',
        array $options = []
    ) {
        if (null === $parentId) {
            $parentId = 'me/skydrive';
        }

        if (is_resource($content)) {
            $stream = $content;
        } else {
            $options = array_merge([
                'stream_back_end' => $this->_streamBackEnd,
            ], $options);

            $stream = $this
                ->_streamOpener
                ->open($options['stream_back_end']);

            if (false === $stream) {
                throw new \Exception('fopen() failed');
            }

            if (false === fwrite($stream, $content)) {
                fclose($stream);
                throw new \Exception('fwrite() failed');
            }

            if (!rewind($stream)) {
                fclose($stream);
                throw new \Exception('rewind() failed');
            }
        }

        $options = array_merge([
            'name_conflict_behavior' => $this->_nameConflictBehavior,
        ], $options);

        $params = $this
            ->_nameConflictBehaviorParameterizer
            ->parameterize([], $options['name_conflict_behavior']);

        $query = http_build_query($params);

        /**
         * @todo some versions of cURL cannot PUT memory streams? See here for a
         * workaround: https://bugs.php.net/bug.php?id=43468
         */
        $file = $this->apiPut(
            $parentId . '/' . $name . ":/content" . "?$query",
            $stream
        );

        // Close the handle only if we opened it within this function.
        if (!is_resource($content)) {
            fclose($stream);
        }

        return new File($this, $file->id, $file);
    }

    /**
     * Fetches a drive item from the current OneDrive account.
     *
     * @param null|string $driveItemId The unique ID of the OneDrive drive item
     *                                 to fetch, or null to fetch the OneDrive
     *                                 root folder. Default: null.
     *
     * @return object The drive item fetched, as a DriveItem instance
     *                referencing to the OneDrive drive item fetched.
     */
    public function fetchObject($driveItemId = null)
    {
        $driveItemId = null !== $driveItemId ? $driveItemId : 'me/skydrive';
        $result      = $this->apiGet($driveItemId);

        if (in_array($result->type, ['folder', 'album'])) {
            return new Folder($this, $driveItemId, $result);
        }

        return new File($this, $driveItemId, $result);
    }

    /**
     * Fetches the root folder from the current OneDrive account.
     *
     * @return Folder The root folder, as a Folder instance referencing to the
     *                OneDrive root folder.
     */
    public function fetchRoot()
    {
        return $this->fetchObject();
    }

    /**
     * Fetches the "Camera Roll" folder from the current OneDrive account.
     *
     * @return Folder The "Camera Roll" folder, as a Folder instance referencing
     *                to the OneDrive "Camera Roll" folder.
     */
    public function fetchCameraRoll()
    {
        return $this->fetchObject('me/skydrive/camera_roll');
    }

    /**
     * Fetches the "Documents" folder from the current OneDrive account.
     *
     * @return Folder The "Documents" folder, as a Folder instance referencing
     *                to the OneDrive "Documents" folder.
     */
    public function fetchDocs()
    {
        return $this->fetchObject('me/skydrive/my_documents');
    }

    /**
     * Fetches the "Pictures" folder from the current OneDrive account.
     *
     * @return Folder The "Pictures" folder, as a Folder instance referencing to
     *                the OneDrive "Pictures" folder.
     */
    public function fetchPics()
    {
        return $this->fetchObject('me/skydrive/my_photos');
    }

    /**
     * Fetches the "Public" folder from the current OneDrive account.
     *
     * @return Folder The "Public" folder, as a Folder instance referencing to
     *                the OneDrive "Public" folder.
     */
    public function fetchPublicDocs()
    {
        return $this->fetchObject('me/skydrive/public_documents');
    }

    /**
     * Fetches the properties of a drive item in the current OneDrive account.
     *
     * @param string $driveItemId The drive item ID.
     *
     * @return object The properties of the drive item fetched.
     */
    public function fetchProperties($driveItemId)
    {
        if (null === $driveItemId) {
            $driveItemId = 'me/skydrive';
        }

        return $this->apiGet($driveItemId);
    }

    /**
     * Fetches the drive items in a folder in the current OneDrive account.
     *
     * @param string $driveItemId The drive item ID.
     *
     * @return array The drive items in the folder fetched, as DriveItem
     *               instances referencing OneDrive drive items.
     */
    public function fetchObjects($driveItemId)
    {
        if (null === $driveItemId) {
            $driveItemId = 'me/skydrive';
        }

        $result     = $this->apiGet($driveItemId . '/files');
        $driveItems = [];

        foreach ($result->data as $data) {
            $driveItem = in_array($data->type, ['folder', 'album']) ?
                new Folder($this, $data->id, $data)
                : new File($this, $data->id, $data);

            $driveItems[] = $driveItem;
        }

        return $driveItems;
    }

    /**
     * Updates the properties of a drive item in the current OneDrive account.
     *
     * @param string       $driveItemId The unique ID of the drive item to
     *                                  update.
     * @param array|object $properties  The properties to update. Default: [].
     * @param bool         $temp        Option to allow save to a temporary file
     *                                  in case of large files.
     *
     * @throws \Exception Thrown on I/O errors.
     */
    public function updateObject($driveItemId, $properties = [], $temp = false)
    {
        $properties = (object) $properties;
        $encoded    = json_encode($properties);
        $stream     = fopen('php://' . ($temp ? 'temp' : 'memory'), 'rw+b');

        if (false === $stream) {
            throw new \Exception('fopen() failed');
        }

        if (false === fwrite($stream, $encoded)) {
            throw new \Exception('fwrite() failed');
        }

        if (!rewind($stream)) {
            throw new \Exception('rewind() failed');
        }

        $this->apiPut($driveItemId, $stream, 'application/json');
    }

    /**
     * Moves a drive item into another folder.
     *
     * @param string      $driveItemId   The unique ID of the drive item to
     *                                   move.
     * @param null|string $destinationId The unique ID of the folder into which
     *                                   to move the drive item, or null to move
     *                                   it to the OneDrive root folder.
     *                                   Default: null.
     */
    public function moveObject($driveItemId, $destinationId = null)
    {
        if (null === $destinationId) {
            $destinationId = 'me/skydrive';
        }

        $this->apiMove($driveItemId, [
            'destination' => $destinationId,
        ]);
    }

    /**
     * Copies a file into another folder. OneDrive does not support copying
     * folders.
     *
     * @param string      $driveItemId   The unique ID of the file to copy.
     * @param null|string $destinationId The unique ID of the folder into which
     *                                   to copy the file, or null to copy it to
     *                                   the OneDrive root folder. Default:
     *                                   null.
     */
    public function copyFile($driveItemId, $destinationId = null)
    {
        if (null === $destinationId) {
            $destinationId = 'me/skydrive';
        }

        $this->apiCopy($driveItemId, [
            'destination' => $destinationId,
        ]);
    }

    /**
     * Deletes a drive item in the current OneDrive account.
     *
     * @param string $driveItemId The unique ID of the drive item to delete.
     */
    public function deleteObject($driveItemId)
    {
        $this->apiDelete($driveItemId);
    }

    /**
     * Fetches the quota of the current OneDrive account.
     *
     * @return object An object with the following properties:
     *                - 'quota' (int) The total space, in bytes.
     *                - 'available' (int) The available space, in bytes.
     */
    public function fetchQuota()
    {
        return $this->apiGet('me/skydrive/quota');
    }

    /**
     * Fetches the account info of the current OneDrive account.
     *
     * @return object An object with the following properties:
     *                - 'id' (string) OneDrive account ID.
     *                - 'first_name' (string) Account owner's first name.
     *                - 'last_name' (string) Account owner's last name.
     *                - 'name' (string) Account owner's full name.
     *                - 'gender' (string) Account owner's gender.
     *                - 'locale' (string) Account owner's locale.
     */
    public function fetchAccountInfo()
    {
        return $this->apiGet('me');
    }

    /**
     * Fetches the recent documents uploaded to the current OneDrive account.
     *
     * @return object An object with the following properties:
     *                - 'data' (array) The list of the recent documents
     *                uploaded.
     */
    public function fetchRecentDocs()
    {
        return $this->apiGet('me/skydrive/recent_docs');
    }

    /**
     * Fetches the drive items shared with the current OneDrive account.
     *
     * @return object An object with the following properties:
     *                - 'data' (array) The list of the shared drive items.
     */
    public function fetchShared()
    {
        return $this->apiGet('me/skydrive/shared');
    }
}
