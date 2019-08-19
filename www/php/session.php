<?php
require_once '../vendor/autoload.php';
require_once '../config.php';

function start_session($sessionrequest) {
    $protocol = explode(':', IRMA_SERVER_URL, 2)[0];

    $jsonsr = json_encode($sessionrequest);

    $api_call = array(
        $protocol => array(
            'method' => 'POST',
            'header' => 'Content-type: application/json\r\n'
                . 'Content-Length: ' . strlen($jsonsr) . '\r\n'
                . 'Authorization: ' . IRMA_SERVER_API_TOKEN . '\r\n',
            'content' => $jsonsr
        )
    );

    $resp = file_get_contents(IRMA_SERVER_URL . '/session', false, stream_context_create($api_call));
    if (! $resp) {
        error();
    }
    return $resp;
}


function start_issuance_session() {
    $randomnum = rand(1,9);
    for($i=0; $i<10; $i++)
        $randomnum .= rand(0,9);

    return start_session([
        "@context" => "https://irma.app/ld/request/issuance/v2",
        "credentials" => [[
            "credential" => "pbdf.pbdf.irmatube",
            "validity" => strtotime("+6 months"),
            "attributes" => [
                "type" => "regular",
                "id" => $randomnum
            ]
        ]]
    ]);
}

function start_verification_session($age = null) {
    $attrs = [
        [
            [ "pbdf.pbdf.irmatube.type" ]
        ]
    ];
    if ($age != null) {
        $attrs[] = [
            ["pbdf.pbdf.ageLimits.over" . $age],
            ["pbdf.nijmegen.ageLimits.over" . $age ],
            ["pbdf.gemeente.personalData.over" . $age ],
        ];
    }
    return start_session([
        "@context" => "https://irma.app/ld/request/disclosure/v2",
        "disclose" => $attrs
    ]);
}

if(!isset($_REQUEST['type']) || empty($_REQUEST['type'])) {
    header("HTTP/1.0 400 Bad Request");
    exit;
}

$type = $_REQUEST['type'];
switch ($type) {
    case "issuance":
        echo start_issuance_session();
        break;
    case "verification":
        $age = $_REQUEST['age'];
        if ($age != null && !ctype_digit($age)) {
            header("HTTP/1.0 400 Bad Request");
            exit;
        }
        echo start_verification_session($age);
        break;
    default:
        header("HTTP/1.0 400 Bad Request");
        exit;
}

exit;
