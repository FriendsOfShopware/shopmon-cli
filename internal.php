<?php

declare(strict_types=1);

namespace ShopmonCli\Internal;

use Composer\InstalledVersions;
use PharData;
use RuntimeException;

use function array_map;
use function chmod;
use function curl_error;
use function curl_exec;
use function curl_getinfo;
use function curl_init;
use function curl_setopt;
use function escapeshellarg;
use function escapeshellcmd;
use function extension_loaded;
use function fclose;
use function file_exists;
use function file_get_contents;
use function file_put_contents;
use function fopen;
use function fprintf;
use function fwrite;
use function implode;
use function ini_get;
use function is_dir;
use function is_resource;
use function mkdir;
use function number_format;
use function php_uname;
use function proc_close;
use function proc_get_status;
use function proc_open;
use function strtolower;
use function unlink;

use const CURLINFO_HTTP_CODE;
use const CURLOPT_FILE;
use const CURLOPT_FOLLOWLOCATION;
use const CURLOPT_NOPROGRESS;
use const CURLOPT_PROGRESSFUNCTION;
use const STDERR;

/**
 * Get the installed shopmon-cli version from Composer metadata.
 *
 * @throws RuntimeException If the version cannot be determined.
 *
 * @return string The package version (e.g., "0.0.4").
 */
function get_version(): string
{
    $version = InstalledVersions::getPrettyVersion('frosh/shopmon-cli');
    if ($version === null) {
        throw new RuntimeException('Could not determine shopmon-cli package version.');
    }

    return $version;
}

/**
 * Detect the CPU architecture from the system.
 *
 * @throws RuntimeException If the architecture is not supported.
 *
 * @return string Normalized architecture (e.g., "amd64", "arm64").
 */
function detect_architecture(): string
{
    $raw = strtolower(php_uname('m'));

    return match ($raw) {
        'x86_64', 'amd64' => 'amd64',
        'arm64', 'aarch64' => 'arm64',
        default => throw new RuntimeException(
            "Unsupported architecture: {$raw}. Pre-built binaries are available for amd64 and arm64.",
        ),
    };
}

/**
 * Detect the operating system and return platform metadata.
 *
 * @param string $architecture Normalized architecture name.
 *
 * @throws RuntimeException If the OS is not supported.
 *
 * @return array{os: string, arch: string}
 */
function detect_platform(string $architecture): array
{
    $os = strtolower(php_uname('s'));

    return match ($os) {
        'linux' => [
            'os' => 'linux',
            'arch' => $architecture,
        ],
        'darwin' => [
            'os' => 'darwin',
            'arch' => $architecture,
        ],
        default => throw new RuntimeException(
            "Unsupported operating system: {$os}. Pre-built binaries are available for Linux and macOS.",
        ),
    };
}

/**
 * Build the archive filename for a given version and platform.
 *
 * @return string Archive name (e.g., "shopmon-cli_0.0.4_linux_amd64").
 */
function build_archive_name(string $version, string $os, string $arch): string
{
    return "shopmon-cli_{$version}_{$os}_{$arch}";
}

/**
 * Build the GitHub release download URL.
 *
 * @return string Full download URL.
 */
function build_download_url(string $version, string $archiveName): string
{
    return "https://github.com/FriendsOfShopware/shopmon-cli/releases/download/{$version}/{$archiveName}.tar.gz";
}

/**
 * Download a file from a URL using the best available method.
 *
 * @throws RuntimeException If the download fails or no download method is available.
 */
function download(string $url, string $destination): void
{
    if (extension_loaded('curl')) {
        namespace\download_with_curl($url, $destination);

        return;
    }

    if (ini_get('allow_url_fopen')) {
        namespace\download_with_fopen($url, $destination);

        return;
    }

    throw new RuntimeException(
        'Unable to download shopmon-cli binary. Either install the PHP curl extension or set allow_url_fopen=1 in php.ini.',
    );
}

/**
 * Download a file using the curl extension with a progress bar.
 *
 * @throws RuntimeException If the download fails or the server returns an error status.
 */
function download_with_curl(string $url, string $destination): void
{
    $ch = curl_init($url);
    $fh = fopen($destination, 'w');
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch, CURLOPT_FILE, $fh);
    curl_setopt($ch, CURLOPT_NOPROGRESS, false);
    curl_setopt($ch, CURLOPT_PROGRESSFUNCTION, function (mixed $_resource, int $dlSize, int $dlNow): int {
        if ($dlSize > 0) {
            $pct = (int) (($dlNow / $dlSize) * 100);
            $dlMb = number_format($dlNow / 1_048_576, 1);
            $totalMb = number_format($dlSize / 1_048_576, 1);
            fprintf(STDERR, "\r  %s / %s MB (%d%%)", $dlMb, $totalMb, $pct);
        }

        return 0;
    });

    $success = curl_exec($ch);
    /** @var int<100, 599> */
    $statusCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $error = curl_error($ch);
    fclose($fh);

    if (!$success || $statusCode >= 400) {
        unlink($destination);

        throw new RuntimeException("Failed to download shopmon-cli binary (HTTP {$statusCode}): {$error}\nURL: {$url}");
    }

    fprintf(STDERR, "\n");
}

/**
 * Download a file using `file_get_contents` (requires `allow_url_fopen`).
 *
 * @throws RuntimeException If the download fails.
 */
function download_with_fopen(string $url, string $destination): void
{
    $contents = file_get_contents($url);
    if ($contents === false) {
        throw new RuntimeException("Failed to download shopmon-cli binary.\nURL: {$url}");
    }

    file_put_contents($destination, $contents);
}

/**
 * Extract a tar.gz archive to a destination directory.
 *
 * @throws RuntimeException If the archive cannot be extracted.
 */
function extract_archive(string $archiveFile, string $destination): void
{
    $phar = new PharData($archiveFile);
    $phar->extractTo($destination);
    unlink($archiveFile);
}

/**
 * Ensure the shopmon-cli binary is available, downloading it if necessary.
 *
 * @param string $version Package version.
 * @param string $archiveName Archive name without extension.
 * @param string $binDir Base directory for storing binaries.
 *
 * @throws RuntimeException If the download, extraction, or binary verification fails.
 *
 * @return string Path to the shopmon-cli executable.
 */
function ensure_binary(string $version, string $archiveName, string $binDir): string
{
    $releaseDir = "{$binDir}/{$version}";
    $executablePath = "{$releaseDir}/shopmon-cli";

    if (file_exists($executablePath)) {
        return $executablePath;
    }

    $archiveFile = "{$releaseDir}/{$archiveName}.tar.gz";
    $url = namespace\build_download_url($version, $archiveName);

    if (!is_dir($releaseDir)) {
        mkdir($releaseDir, 0o755, true);
    }

    fprintf(STDERR, "Downloading shopmon-cli %s for %s...\n", $version, $archiveName);
    namespace\download($url, $archiveFile);
    fprintf(STDERR, "Downloaded.\n");

    namespace\extract_archive($archiveFile, $releaseDir);

    if (!file_exists($executablePath)) {
        throw new RuntimeException("Expected binary not found after extraction at {$executablePath}");
    }

    chmod($executablePath, 0o755);

    return $executablePath;
}

/**
 * Execute the shopmon-cli binary, forwarding stdin/stdout/stderr.
 *
 * @param string $executablePath Path to the shopmon-cli binary.
 * @param list<string> $args Command-line arguments to pass.
 *
 * @return never
 */
function execute(string $executablePath, array $args): never
{
    $command = escapeshellcmd($executablePath);
    if ($args !== []) {
        $command .= ' ' . implode(' ', array_map(escapeshellarg(...), $args));
    }

    $pipes = [];
    $process = @proc_open(
        $command,
        [
            0 => ['file', 'php://stdin', 'r'],
            1 => ['file', 'php://stdout', 'w'],
            2 => ['file', 'php://stderr', 'w'],
        ],
        $pipes,
    );

    if (!is_resource($process)) {
        fwrite(STDERR, "Error: Unable to start shopmon-cli process.\n");
        exit(1);
    }

    do {
        $status = proc_get_status($process);
    } while ($status['running']);

    $exitCode = $status['exitcode'];
    proc_close($process);

    exit($exitCode);
}
