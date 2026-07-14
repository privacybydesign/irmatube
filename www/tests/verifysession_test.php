<?php
// Regression tests for the error/failure branches of www/php/verifysession.php.
//
// The bug (issue #28): every error/failure branch called echo BEFORE
// http_response_code()/header(), so once output_buffering was off the status
// line could not be sent any more ("headers already sent") and the client got
// a 200 with a plain-text body instead of the intended 400/403 JSON.
//
// Each case runs the real endpoint in a subprocess with output_buffering
// explicitly OFF -- the setting under which the bug surfaced -- and asserts the
// intended HTTP status AND a JSON body. These tests fail against the old
// ordering and pass after the fix.
//
// No test framework: run it directly with any PHP CLI (composer deps must be
// installed in www/):
//   php www/tests/verifysession_test.php

require __DIR__ . '/../vendor/autoload.php';

use Firebase\JWT\JWT;

$phpBin  = PHP_BINARY;
$phpDir  = realpath(__DIR__ . '/../php');
$prepend = __DIR__ . '/prepend_input_mock.php';
$dataDir = realpath(__DIR__ . '/../../data') . '/';
$pubkey  = $dataDir . 'demo-publickey.pem';
$privkey = $dataDir . 'demo-privkey.pem';
$credId  = 'irma-demo.IRMATube.member';

// Signed RS256 tokens. Both decode successfully; they differ only in whether
// the disclosed attributes prove IRMATube membership.
$sign = function (array $disclosed) use ($privkey) {
    return JWT::encode(['disclosed' => $disclosed], file_get_contents($privkey), 'RS256');
};
$nonMemberToken = $sign([[(object) ['id' => 'irma-demo.other.attr', 'rawvalue' => 'x']]]);
$memberToken    = $sign([[(object) ['id' => $credId . '.type', 'rawvalue' => 'regular']]]);

// Extra env for the branches that get past the input checks.
$verifyEnv = [
    'ROOT_DIR'               => $dataDir,
    'IRMA_SERVER_PUBLICKEY'  => $pubkey,
    'IRMATUBE_CREDENTIAL_ID' => $credId,
];

/**
 * Run verifysession.php in a subprocess with output_buffering off and the given
 * JSON request body. Returns [status_code, stdout_body].
 */
function run_endpoint($phpBin, $phpDir, $prepend, $body, array $env)
{
    $cmd = escapeshellarg($phpBin)
        . ' -d output_buffering=0'
        . ' -d display_errors=stderr'
        . ' -d auto_prepend_file=' . escapeshellarg($prepend)
        . ' ' . escapeshellarg($phpDir . '/verifysession.php');

    $childEnv = array_merge(getenv(), $env, ['YIVI_TEST_BODY' => $body]);
    $descriptors = [1 => ['pipe', 'w'], 2 => ['pipe', 'w']];
    $proc = proc_open($cmd, $descriptors, $pipes, $phpDir, $childEnv);
    if (!is_resource($proc)) {
        throw new RuntimeException('failed to start php subprocess');
    }
    $stdout = stream_get_contents($pipes[1]);
    fclose($pipes[1]);
    $stderr = stream_get_contents($pipes[2]);
    fclose($pipes[2]);
    proc_close($proc);

    $code = null;
    if (preg_match('/__HTTP_CODE__=(\d+)/', $stderr, $m)) {
        $code = (int) $m[1];
    }
    return [$code, $stdout];
}

$failures = 0;
$run = 0;

function check($label, $cond, &$failures, &$run)
{
    $run++;
    if ($cond) {
        echo "  PASS: $label\n";
    } else {
        $failures++;
        echo "  FAIL: $label\n";
    }
}

// --- missing token -> 400 + JSON ------------------------------------------
[$code, $body] = run_endpoint($phpBin, $phpDir, $prepend, json_encode(new stdClass()), []);
$json = json_decode($body, true);
echo "[missing token] status=$code body=$body\n";
check('missing token returns 400', $code === 400, $failures, $run);
check('missing token body is JSON', $json !== null, $failures, $run);
check('missing token body success=false', is_array($json) && ($json['success'] ?? null) === false, $failures, $run);

// --- missing videoid -> 400 + JSON ----------------------------------------
[$code, $body] = run_endpoint($phpBin, $phpDir, $prepend, json_encode(['token' => 'anything']), []);
$json = json_decode($body, true);
echo "[missing videoid] status=$code body=$body\n";
check('missing videoid returns 400', $code === 400, $failures, $run);
check('missing videoid body is JSON', $json !== null, $failures, $run);
check('missing videoid body success=false', is_array($json) && ($json['success'] ?? null) === false, $failures, $run);

// --- verification fails (valid JWT, not a member) -> 403 + JSON -----------
[$code, $body] = run_endpoint(
    $phpBin, $phpDir, $prepend,
    json_encode(['token' => $nonMemberToken, 'videoid' => 'up']),
    $verifyEnv
);
$json = json_decode($body, true);
echo "[verification failed] status=$code body=$body\n";
check('verification failure returns 403', $code === 403, $failures, $run);
check('verification failure body is JSON', $json !== null, $failures, $run);
check('verification failure body success=false', is_array($json) && ($json['success'] ?? null) === false, $failures, $run);

// --- happy path (member, no age restriction) -> 200 + JSON ----------------
[$code, $body] = run_endpoint(
    $phpBin, $phpDir, $prepend,
    json_encode(['token' => $memberToken, 'videoid' => 'up']),
    $verifyEnv
);
$json = json_decode($body, true);
echo "[happy path] status=$code body=$body\n";
check('happy path returns 200', $code === 200, $failures, $run);
check('happy path body is JSON', $json !== null, $failures, $run);
check('happy path body success=true', is_array($json) && ($json['success'] ?? null) === true, $failures, $run);
check('happy path returns youtubeId', is_array($json) && !empty($json['youtubeId']), $failures, $run);

echo "\n$run checks, $failures failure(s)\n";
exit($failures === 0 ? 0 : 1);
