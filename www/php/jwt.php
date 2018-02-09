<?php
require_once '../vendor/autoload.php';
require_once '../config.php';
use \Firebase\JWT\JWT;

// Generate a disclosure or issuance JWT

function get_jwt_key() {
    $path = "file://" . ROOT_DIR . JWT_PRIVATEKEY;
    $pk = openssl_pkey_get_private($path);
    if ($pk === false)
        throw new Exception("Failed to load signing key " . $path);
    return $pk;
}

function get_issuance_jwt() {
    $pk = get_jwt_key();

    $randomnum = rand(1,9);
    for($i=0; $i<10; $i++)
        $randomnum .= rand(0,9);

    $iprequest = [
        "sub" => "issue_request",
        "iss" => "IRMATube",
        "iat" => time(),
        "iprequest" => [
            "timeout" => 300,
            "request" => [
                "credentials" => [
                    [
                        "credential" => "pbdf.pbdf.irmatube",
                        "attributes" => [
                            "type" => "regular",
                            "id" => $randomnum
                        ]
                    ]
                ]
            ]
        ]
    ];

    return JWT::encode($iprequest, $pk, "RS256", "irmatube");
}

function get_verification_jwt($age = null) {
    $pk = get_jwt_key();

    $sprequest = [
        "sub" => "verification_request",
        "iss" => "IRMATube",
        "iat" => time(),
        "sprequest" => [
            "timeout" => 300,
            "request" => [
                "content" => [
                    [
                        "label" => "Membership",
                        "attributes" => [ "pbdf.pbdf.irmatube.type" ]
                    ]
                ]
            ]
        ]
    ];

    if ($age != null) {
        $sprequest["sprequest"]["request"]["content"][] = [
            "label" => "iDIN over " . $age,
            "attributes" => [ "pbdf.pbdf.ageLimits.over" . $age ]
        ];
    }

    return JWT::encode($sprequest, $pk, "RS256", "irmatube");
}

if(!isset($_REQUEST['type']) || empty($_REQUEST['type'])) {
    header("HTTP/1.0 400 Bad Request");
    exit;
}

$type = $_REQUEST['type'];
switch ($type) {
    case "issuance":
        echo get_issuance_jwt();
        break;
    case "verification":
        $age = $_REQUEST['age'];
        if ($age != null && !ctype_digit($age)) {
            header("HTTP/1.0 400 Bad Request");
            exit;
        }
        echo get_verification_jwt($age);
        break;
    default:
        header("HTTP/1.0 400 Bad Request");
        exit;
}

exit;

?>
