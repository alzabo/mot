# mot: a qBittorrent CLI

`mot` is a command-line interface (CLI) for interacting with qBittorrent in
the UNIX tradition.

## Features

* **Add torrents:** Add torrents from files or URLs.
* **List torrents:** View the status of your torrents.
* **Pause/Resume/Delete torrents:** Control downloads and seeding.
* **Set download location:** Specify where torrents should be saved.
* **Category management**: Add, remove and list categories.
* **Tag management**: Add, remove and list tags.
* **WebUI alternative**: For headless servers or terminal lovers.
* **Cross Platform**: Supports Linux, macOS, and Windows.

## Installation

### From Source

```sh
go install github.com/alzabo/mot@latest
```

## Setup

### Prerequisites

* qBittorrent Web UI enabled.

### Authentication

The qBittorrent Web UI username, password and URL must be supplied to
allow mot to connect. These configuration options may be set using one of
the methods described below.

> Note: Replace "admin" and "password" with your qBittorrent Web UI credentials and
> the URL with the correct value.

#### CLI args

```sh
mot --username admin --password password --url http://localhost:8080 get torrents
```

#### Environment variables

```bash
export MOT_USERNAME=admin
export MOT_PASSWORD=password
export MOT_URL=http://localhost:8080
mot get torrents
```

### Config file

By default, mot looks for its configuration in ~/.mot.yaml. A different
location may be specified by setting the `--config` flag or the `MOT_CONFIG`
environment var.

Example ~/.mot.yaml:

```yaml
username: "admin"
password: "password"
url: "http://localhost:8080"
```

## Usage

### Basic Commands

Add a torrent from a file:

```bash
mot add torrent -f path/to/torrent.torrent
```

Add a torrent from a magnet link:

```bash
mot add torrent --url magnet:?xt=urn:btih:...
```

List all torrents:

```bash
mot get torrents
```

Pause a torrent:

```bash
mot pause <torrent_hash>
```

Resume a torrent:

```bash
mot resume <torrent_hash>
```

Delete a torrent (and optionally data):

```bash
mot delete <torrent_hash> [--delete-data]
```

Help
For detailed usage information, use the --help flag:

```bash
mot --help
mot <command> --help
```
