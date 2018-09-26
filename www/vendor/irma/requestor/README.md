# irmarequestor

`irmarequestor` can create signed JWT disclosure or issuance requests to send to an [`irma_api_server`](https://github.com/privacybydesign/irma_api_server), for example using [`irma_js`](https://github.com/privacybydesign/irma_js). It can be used to start IRMA attribute disclosure or issuance sessions with an [IRMA mobile app](https://github.com/privacybydesign/irma_mobile).

## Install using composer

```
composer require irma/requestor
```

## Use

Create an instance:
```php
include "vendor/autoload.php";
$requestor = new \IRMA\Requestor("Name", "id", "privatekey.pem");
```

The first two parameters identifiy your application to the [IRMA app user](https://github.com/privacybydesign/irma_mobile) and to the `irma_api_server`, respectively. The private key with which to sign the JWT must be stored in PEM format at the path specified by the third parameter.

Create a verification request:
```php
$jwt = $requestor->getVerificationJwt([
	[
		"label" => "IRMATube member type",
		"attributes" => [ "irma-demo.IRMATube.member.type" ]
	]
]);
```

Create an issuance request:
```php
$jwt = $requestor->getIssuanceJwt([
	[
		"credential" => "irma-demo.IRMATube.member",
		"attributes" => [ "type" => "regular", "id" => "2" ]
	]
]);
```
