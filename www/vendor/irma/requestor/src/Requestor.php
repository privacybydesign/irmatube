<?php
namespace IRMA;
use \Exception;
use \Firebase\JWT\JWT;

/**
 * Class Requestor creates signed IRMA issuance or verification requests.
 * @package IRMA
 */
class Requestor {
	private $name, $id, $privatekey;

	/**
	 * Requestor constructor.
	 * @param $name string Name to show to the IRMA user
	 * @param $id string Identifier under which we are known to the irma_api_server
	 * @param $privatekey string Path to the RSA private key file in PEM encoding with which to sign the JWTs
	 */
	public function __construct($name, $id, $privatekey) {
		$this->name = $name;
		$this->id = $id;
		$this->privatekey = $privatekey;
	}

	private function getJwtPrivateKey() {
		if (!function_exists("openssl_pkey_get_private"))
			throw new Exception("openssl functions not available");
		$realpath = realpath($this->privatekey);
		if ($realpath === false)
			throw new Exception("No readable private key at " . $this->privatekey);

		$pk = openssl_pkey_get_private("file://" . $realpath);
		if ($pk === false)
			throw new Exception("Failed to load signing key: " . openssl_error_string());
		return $pk;
	}

	/**
	 * Create a signed IRMA attribute verification request.
	 * @param $attributes array Attributes to verify
	 * @return string Signed JWT to POST to a irma_api_server
	 * @throws Exception If $attributes does not have the right format or the private key could not be read
	 */
	public function getVerificationJwt($attributes) {
		if (!is_array($attributes) || count($attributes) === 0)
			throw new Exception("Argument not a nonempty array of disjunctions");
		foreach ($attributes as $disjunction) {
			if (!isset($disjunction["label"]))
				throw new Exception("Disjunction has no label");
			if (!isset($disjunction["attributes"]) || !is_array($disjunction["attributes"]) || count($disjunction["attributes"]) === 0)
				throw new Exception("Disjunction has no attributes");
		}

		$request = [
			"iat" => time(),
			"iss" => $this->name,
			"sub" => "verification_request",
			"sprequest" => [
				"validity" => 60,
				"request" => [
					"content" => $attributes
				]
			]
		];

		return JWT::encode($request, $this->getJwtPrivateKey(), "RS256", $this->id);
	}

	/**
	 * Create a signed IRMA issuance request.
	 * @param $credentials array Credentials to issue
	 * @return string Signed JWT to POST to a irma_api_server
	 * @throws Exception If $attributes does not have the right format or the private key could not be read
	 */
	function getIssuanceJwt($credentials) {
		if (!is_array($credentials) || count($credentials) === 0)
			throw new Exception("Argument not a nonempty array of credentials");
		foreach ($credentials as $cred) {
			if (!isset($cred["credential"]))
				throw new Exception("Credential has no identifier");
			if (!isset($cred["attributes"]) || !is_array($cred["attributes"]) || count($cred["attributes"]) === 0)
				throw new Exception("Credential has no attributes");
		}

		$request = [
			"iat" => time(),
			"iss" => $this->name,
			"sub" => "issue_request",
			"iprequest" => [
				"timeout" => 300,
				"request" => [
					"credentials" => $credentials
				]
			]
		];

		return JWT::encode($request, $this->getJwtPrivateKey(), "RS256", $this->id);
	}
}
