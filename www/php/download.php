<?php
require_once '../vendor/autoload.php';
require_once '../config.php';
require_once 'range_serve.php';
use \Firebase\JWT\JWT;
use \Firebase\JWT\Key;

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

$jwt_pk = file_get_contents(ROOT_DIR . IRMA_SERVER_PUBLICKEY);

$token = $_REQUEST['token'];

// We want the movies to continue playing for an hour
JWT::$leeway = 60 * 60;
try {
    $decoded = JWT::decode($token, new Key($jwt_pk, 'RS256'));
} catch (Exception $e) {
    header("HTTP/1.0 403 Forbidden");
    exit;
}
$disclosed = (array) $decoded->disclosed;

function isMember($disclosed) {
    $member_key = IRMATUBE_CREDENTIAL_ID . ".type";
    foreach ($disclosed as $con) {
        foreach ($con as $attr) {
            if ($attr->id == $member_key) {
                return $attr->rawvalue === "regular" || $attr->rawvalue === "premium";
            }
        }
    }

    return false;
}

function isAgeAllowed($file_path, $disclosed) {
    if(!isset($file_path) || !isset($disclosed)) {
        return false;
    }

    $age_file = file($file_path . ".access");
    $age_restriction = intval(trim($age_file[0]));

    if($age_restriction == 0) {
        // Movie has no age restriction
        return true;
    }

    $age_key_pbdf = "pbdf.pbdf.ageLimits.over" . $age_restriction;
    $age_key_nijmegen = "pbdf.nijmegen.ageLimits.over" . $age_restriction;
    $age_key_gemeente = "pbdf.gemeente.personalData.over" . $age_restriction;
    $age_key_idcard = "pbdf.pilot-amsterdam.idcard.over" . $age_restriction;
    $age_key_passport = "pbdf.pilot-amsterdam.passport.over" . $age_restriction;

    foreach ($disclosed as $con) {
        foreach ($con as $attr) {
            if ($attr->id == $age_key_pbdf
                || $attr->id == $age_key_nijmegen
                || $attr->id == $age_key_gemeente
                || $attr->id == $age_key_idcard
                || $attr->id == $age_key_passport
            ) {
                return strtolower($attr->rawvalue) == "yes" || strtolower($attr->rawvalue) == "ja";
            }
        }
    }

    return false;
}

if( isAgeAllowed($file_path, $disclosed) && isMember($disclosed) ) {
    range_serve($base_path);
} else {
    header("HTTP/1.0 403 Forbidden");
    exit;
}

exit;

?>
