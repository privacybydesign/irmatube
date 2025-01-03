<?php

define('ROOT_DIR', getenv('ROOT_DIR') ?: __DIR__ . '/../data/');

// URL and JWT public key of your IRMA server, also include your access token if enabled on your server
define('IRMA_SERVER_URL',getenv('IRMA_SERVER_URL') ?: 'http://localhost:8088');
define('IRMA_SERVER_API_TOKEN', getenv('IRMA_SERVER_API_TOKEN')?: '');
define('IRMATUBE_CREDENTIAL_ID',getenv('IRMATUBE_CREDENTIAL_ID') ?: 'irma-demo.IRMATube.member');
define('IRMA_SERVER_PUBLICKEY',getenv('IRMA_SERVER_PUBLICKEY')?: '/data/demo-publickey.pem');
$language = 'en';

$movies = array("django", "elysium", "oblivion", "olympus-has-fallen",
    "oz-the-great-and-powerful", "planes", "up", "were-the-millers");
