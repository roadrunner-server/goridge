<?php
/**
 * Dead simple, high performance, drop-in bridge to Golang RPC with zero dependencies
 *
 * @author Wolfy-J
 */

namespace Spiral\Tests\Prototypes;

use PHPUnit\Framework\TestCase;
use Symfony\Component\Process\Process;

abstract class ProcessTest extends TestCase
{
    const FORCE_BUILD = false;
    const GO_APP      = "socket4bytes";
    const GO_ARGS     = "";

    /** @var Process */
    protected $process;

    public function setUp()
    {
        $dir = dirname(dirname(__DIR__));
        $file = static::GO_APP;

        if (strtoupper(substr(PHP_OS, 0, 3)) === 'WIN') {
            if (self::FORCE_BUILD || !file_exists($dir . "/{$file}.exe")) {
                $build = new Process("go build {$file}.go", $dir);
                $build->mustRun();
            }

            //enable in firewall
            $this->process = new Process("{$file}.exe " . static::GO_ARGS, $dir);
        } else {
            if (self::FORCE_BUILD || !file_exists($dir . "/{$file}")) {
                $build = new Process("go build {$file}", $dir);
                $build->mustRun();
            }

            $this->process = new Process("go run {$file} " . static::GO_ARGS, $dir);
        }

        $this->process->start();
    }

    public function tearDown()
    {
        if ($this->process->isRunning()) {
            $this->process->signal(0);
        }
    }

    public function testConnected()
    {
        $this->assertTrue($this->process->isStarted());
        $this->assertTrue($this->process->isRunning());
    }

    public function testDieAndRestart()
    {
        $this->assertTrue($this->process->isRunning());
        $this->process->signal(0);
        $this->assertFalse($this->process->isRunning());

        $this->process->start();
        $this->assertTrue($this->process->isRunning());
    }
}