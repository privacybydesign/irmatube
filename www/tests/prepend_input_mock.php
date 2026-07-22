<?php
// Test-only helper (loaded via -d auto_prepend_file). It lets verifysession.php
// be exercised from the CLI, where php://input is normally empty:
//  * feeds the request body from the YIVI_TEST_BODY env var into php://input;
//  * reports the final HTTP status code on STDERR so the test runner can read it.

class MockPhpInputStream
{
    public $context;
    private $data = '';
    private $pos = 0;

    public function stream_open($path, $mode, $options, &$opened_path)
    {
        // Only php://input is mocked; nothing else should reach this wrapper.
        if ($path !== 'php://input') {
            return false;
        }
        $this->data = getenv('YIVI_TEST_BODY');
        if ($this->data === false) {
            $this->data = '';
        }
        $this->pos = 0;
        return true;
    }

    public function stream_read($count)
    {
        $chunk = substr($this->data, $this->pos, $count);
        $this->pos += strlen($chunk);
        return $chunk;
    }

    public function stream_eof()
    {
        return $this->pos >= strlen($this->data);
    }

    public function stream_stat()
    {
        return [];
    }

    public function stream_close()
    {
    }
}

stream_wrapper_unregister('php');
stream_wrapper_register('php', 'MockPhpInputStream');

register_shutdown_function(function () {
    $code = http_response_code();
    // http_response_code() returns false when no code was set (implicit 200).
    if ($code === false) {
        $code = 200;
    }
    fwrite(STDERR, "\n__HTTP_CODE__=" . $code . "\n");
});
