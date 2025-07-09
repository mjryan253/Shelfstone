# Shelfstone

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)

A modern, self-hosted ebook server designed for simplicity and a beautiful user experience. Think Plex, but for your books.

Shelfstone is an open-source project built to be a user-friendly alternative to Calibre's web interface. It runs in Docker, sets up in minutes, and provides a clean, modern interface for you to access your entire ebook library from any device with a web browser.

![Shelfstone Placeholder Screenshot](https://raw.githubusercontent.com/mjryan253/Shelfstone/main/placeholder.png)
*(Note: This is a placeholder image. Add a real screenshot of the UI once it's developed.)*

---

## The Vision

The goal of Shelfstone is to be the best way to manage and read your personal ebook collection. We focus on three core principles:

1.  **Ease of Setup:** Get up and running in under 5 minutes with a simple `docker-compose.yaml`. No complex configuration or database administration required.
2.  **Modern Interface:** A fast, clean, and responsive web UI for Browse your library, viewing book details, and reading directly in the browser.
3.  **Rock-Solid Backend:** Built on Go for a lightweight, high-performance, and stable foundation.

## Key Features (Roadmap)

Shelfstone is currently in the early stages of development. Here is the planned feature set:

* **📚 Automatic Library Scanning:** Drop your ebook files (`.epub`, `.mobi`, `.azw3`, etc.) into your library folder, and Shelfstone will automatically import them.
* **✍️ Metadata & Cover Extraction:** Uses the powerful Calibre backend tools to automatically fetch metadata and cover art.
* **🌐 In-Browser Reading:** Read your books anywhere with a built-in, responsive EPUB web reader.
* **🗂️ Smart Organization:** Automatically organizes your library into a clean folder structure by author and series.
* **👥 User Management (Future):** Create separate accounts for family or friends.
* **🔄 Reading Progress Sync (Future):** Pick up right where you left off, on any device.

---

## Getting Started (Quick Start)

The preferred—and easiest—way to run Shelfstone is with Docker.

**Prerequisites:**
* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/install/)

**Instructions:**

1.  **Create a `docker-compose.yaml` file:**

    ```yaml
    version: '3.8'

    services:
      shelfstone:
        image: ghcr.io/mjryan253/shelfstone:latest # This image does not exist yet
        container_name: shelfstone
        ports:
          - "8080:8080"
        volumes:
          - ./config:/config   # Stores database and config files
          - ./books:/books     # Your ebook library
        restart: unless-stopped
    ```

2.  **Create your library folder:**

    ```bash
    mkdir books
    mkdir config
    ```
    Place your ebook files inside the newly created `books` directory.

3.  **Start the server:**

    ```bash
    docker-compose up -d
    ```

4.  **Access the web UI:** Open your browser and navigate to `http://localhost:8080`.

---

## Technology Stack

Shelfstone is built with modern, performant technologies:

* **Backend:** [Go (Golang)](https://go.dev/) with the [Gin](https://gin-gonic.com/) framework.
* **Database:** [SQLite](https://www.sqlite.org/index.html) for simplicity and portability.
* **Ebook Processing:** Leverages the battle-tested [Calibre command-line tools](https://manual.calibre-ebook.com/generated/en/cli-index.html).
* **Frontend (Planned):** [SvelteKit](https://kit.svelte.dev/) for a fast and reactive user interface.
* **Containerization:** [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/).

## Contributing

This is an open-source project, and contributions are welcome! Whether it's reporting a bug, submitting a feature request, or writing code, please check out the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines.

We are actively looking for developers, especially those with experience in **Go** and **Svelte**.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.