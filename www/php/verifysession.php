<?php
require_once '../vendor/autoload.php';
require_once '../config.php';
use \Firebase\JWT\JWT;
use \Firebase\JWT\Key;

// This script checks JWT validity and presense of irmatube membership and checks age restriction provided in videoid.json to see if ageover.agerestiction is present in the disclosed attributes
// only if these are true, it returns the youtubeId of the video


// token and videoid are passed in the body of the request
$json = file_get_contents('php://input');
$data = json_decode($json, true);


if (!isset($data['token']) || empty($data['token'])) {
    echo "No token provided";
    header("HTTP/1.0 400 Bad Request");
    exit;
}

if (!isset($data['videoid']) || empty($data['videoid'])) {
    echo "No videoid provided";
    header("HTTP/1.0 400 Bad Request");
    exit;
}

$jwt_pk = file_get_contents(ROOT_DIR . IRMA_SERVER_PUBLICKEY);

$token = $data['token'];
$videoid = $data['videoid'];

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

function getAgeRestriction(string $videoid){
    $json = file_get_contents(ROOT_DIR . "videos/" . $videoid . ".json");
    $data = json_decode($json);
    return $data->ageRestriction;
}

function getYTid($videoid){
    $json = file_get_contents(ROOT_DIR . "videos/" . $videoid . ".json");
    $data = json_decode($json);
    return $data->youtubeId;

}

function isAgeAllowed($videoid, $disclosed) {
    $age_restriction= getAgeRestriction($videoid);

    //if movie has no age restriction, no need to check for age credentials in the disclosed attributes
    if ($age_restriction == 0) {
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


if( isAgeAllowed($videoid, $disclosed) && isMember($disclosed) ) {
    $youtubeId = getYTid($videoid);
    echo json_encode(['success' => true, 'youtubeId' => $youtubeId]);
    http_response_code(200);
} else {
    echo json_encode(['success' => false]);
    header("HTTP/1.0 403 Forbidden");

    exit;
}

exit;

?>
