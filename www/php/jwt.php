<?php
require_once '../vendor/autoload.php';
require_once '../config.php';
use \Firebase\JWT\JWT;
use \IRMA\Requestor;

function get_issuance_jwt($requestor) {
    $randomnum = rand(1,9);
    for($i=0; $i<10; $i++)
        $randomnum .= rand(0,9);

    return $requestor->getIssuanceJwt([
        [
            "credential" => "pbdf.pbdf.irmatube",
            "attributes" => [
                "type" => "regular",
                "id" => $randomnum
            ]
        ]
    ]);
}

function get_verification_jwt($requestor, $age = null) {
    $attrs = [
        [
            "label" => "Membership",
            "attributes" => [ "pbdf.pbdf.irmatube.type" ]
        ]
    ];
    if ($age != null) {
        $attrs[] = [
            "label" => "iDIN over " . $age,
            "attributes" => [ "pbdf.pbdf.ageLimits.over" . $age, "pbdf.nijmegen.ageLimits.over" . $age ]
        ];
    }
    return $requestor->getVerificationJwt($attrs);
}

if(!isset($_REQUEST['type']) || empty($_REQUEST['type'])) {
    header("HTTP/1.0 400 Bad Request");
    exit;
}

$requestor = new Requestor("IRMATube", "irmatube", ROOT_DIR . JWT_PRIVATEKEY);
$type = $_REQUEST['type'];
switch ($type) {
    case "issuance":
        echo get_issuance_jwt($requestor);
        break;
    case "verification":
        $age = $_REQUEST['age'];
        if ($age != null && !ctype_digit($age)) {
            header("HTTP/1.0 400 Bad Request");
            exit;
        }
        echo get_verification_jwt($requestor, $age);
        break;
    default:
        header("HTTP/1.0 400 Bad Request");
        exit;
}

exit;

?>
