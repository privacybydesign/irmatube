<?php
require_once '../vendor/autoload.php';
require_once '../config.php';
require_once 'range_serve.php';
use \Firebase\JWT\JWT;

// Stream the specified file, if it exists, and if the IRMA JWT:
// - is validly signed by the API server;
// - specifies that the user had the membership attribute and if necessary, the applicable age attribute.

if(!isset($_REQUEST['file']) || empty($_REQUEST['file']))
{
    header("HTTP/1.0 400 Bad Request");
    exit;
}

if(!isset($_REQUEST['token']) || empty($_REQUEST['token']))
{
    header("HTTP/1.0 400 Bad Request");
    exit;
}

// sanitize the file request, keep just the name and extension
$file_path  = $_REQUEST['file'];
$path_parts = pathinfo($file_path);
$base_name  = $path_parts['basename'];
$file_name  = $path_parts['filename'];
$file_ext   = $path_parts['extension'];

// Limit ourselves the an explicit set of filenames, just a safety measure
if (!in_array($file_name, $movies)) {
    header("HTTP/1.0 404 Not Found");
    exit;
}

$base_path = ROOT_DIR . "videos/" . $base_name;
$file_path = ROOT_DIR . "videos/" . $file_name;

$jwt_pk = file_get_contents(ROOT_DIR . API_SERVER_PUBLICKEY);

$token = $_REQUEST['token'];

// We want the movies to continue playing for an hour
JWT::$leeway = 60 * 60;
try {
    $decoded = JWT::decode($token, $jwt_pk, array('RS256'));
} catch (Exception $e) {
    header("HTTP/1.0 403 Forbidden");
    exit;
}
$attributes = (array) $decoded->attributes;

function isMember($attributes) {
    $member_key = "pbdf.pbdf.irmatube.type";
    if (array_key_exists($member_key, $attributes) && $attributes[$member_key] == "regular")
        return true;
    return false;
}

function isAgeAllowed($file_path, $attributes) {
    if(!isset($file_path) || !isset($attributes)) {
        return false;
    }

    $age_file = file($file_path . ".access");
    $age_restriction = intval(trim($age_file[0]));

    if($age_restriction == 0) {
        // Movie has no age restriction
        return true;
    }

    $pbdf_age_key = "pbdf.pbdf.ageLimits.over" . $age_restriction;
    $nijmegen_age_key = "pbdf.pbdf.nijmegen.over" . $age_restriction;

    return (array_key_exists($age_key_pbdf, $attributes) && $attributes[$age_key_pbdf] === "yes") ||
        (array_key_exists($age_key_nijmegen, $attributes) && $attributes[$age_key_nijmegen] === "yes");
}

if( isAgeAllowed($file_path, $attributes) && isMember($attributes) ) {
    range_serve($base_path);
} else {
    header("HTTP/1.0 403 Forbidden");
    exit;
}

exit;

?>
