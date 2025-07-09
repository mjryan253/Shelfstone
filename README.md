# ShelfStone: Your Personal Ebook Server

ShelfStone is a self-hosted ebook server designed for ease of setup and a clean, modern user experience. It aims to be a lightweight, Docker-first alternative to more complex solutions, allowing you to effortlessly manage and access your ebook library.

**Current Status: Phase 1 Complete (Foundation)**

The core backend services, database setup, and Docker containerization are complete. ShelfStone is ready for further development of its library management and user interface features.

## Core Principles

*   **Ease of Setup:** Get up and running quickly with Docker.
*   **Clean User Experience:** A modern web interface for reading and management (Future Phase).
*   **Leveraging Calibre:** Utilizes the powerful and mature Calibre command-line tools for all backend ebook processing (metadata, conversion), wrapped in a new Go-based API and prepared for a modern UI.
*   **Performance:** Built with Go (Gin) for a fast and efficient backend.
*   **Simplicity:** SQLite database for minimal external dependencies.
*   **Docker First:** Designed for easy deployment and management via Docker containers.

## ShelfStone and Calibre

ShelfStone aims to provide an updated UI/UX and an easy-to-deploy Docker container solution that leverages the proven capabilities of the underlying Calibre command-line interface (CLI). Instead of reinventing ebook processing, ShelfStone wraps Calibre's `ebook-meta` and `ebook-convert` tools, focusing on providing a modern API and (in future phases) a user-friendly web interface. This means you get Calibre's extensive format support and robust conversion engine, managed through a streamlined, containerized application.

## Features (Phase 1 - Foundation)

*   **Go Backend (Gin Framework):** A lightweight and performant API foundation.
    *   Includes a `/api/health` endpoint for monitoring.
*   **SQLite Database (GORM):** Simple, file-based database for storing book metadata, authors, series, and user information.
    *   Automatic schema migration on startup.
*   **Calibre CLI Integration:** Direct wrapping of Calibre's `ebook-meta` (for metadata extraction) and `ebook-convert` (for format conversion) CLI tools. This ensures comprehensive ebook format support.
*   **Dockerized Deployment:**
    *   **Multi-stage `Dockerfile`:** Creates a minimal, efficient production image containing the Go application and the necessary Calibre CLI tools.
    *   **`docker-compose.yaml`:** Simplifies setup and manages:
        *   The main application service.
        *   Persistent volumes for your ebook library (`/books`), application configuration and database (`/config`), and generated cache (`/cache`).
        *   Port mapping (default `8080:8080`).
*   **Basic Project Structure:** Ready for implementation of Phase 2 (MVP Features) including library scanning and core API endpoints.

## Getting Started (Using Docker)

**Prerequisites:**

*   Docker installed (https://www.docker.com/get-started)
*   Docker Compose installed (usually included with Docker Desktop)

**Setup:**

1.  **Clone the Repository (or download the source code):**
    ```bash
    # git clone <repository_url> # Replace with actual URL when available
    # cd shelfstone
    ```

2.  **Prepare Host Directories for Books and Configuration:**
    ShelfStone needs directories on your host machine to store your ebook files and its own configuration (like the database).

    Create these directories if they don't exist. For example:
    ```bash
    mkdir -p my_shelfstone_library/books
    mkdir -p my_shelfstone_library/config
    ```
    *   The `my_shelfstone_library/books` directory is where you will place your ebook files (`.epub`, `.mobi`, `.azw3`, etc.).
    *   The `my_shelfstone_library/config` directory will store ShelfStone's database and any other configuration files.

3.  **Configure `docker-compose.yaml`:**
    Open the `docker-compose.yaml` file in the project root. You **must** update the volume paths to point to the directories you created in the previous step.

    Find this section:
    ```yaml
    volumes:
      - ./example_library/books:/books
      - ./example_library/config:/config
      - libreplex_cache:/cache # Docker-managed volume for cache
    ```
    Change it to match your paths:
    ```yaml
    volumes:
      - /path/to/your/my_shelfstone_library/books:/books  # <-- UPDATE THIS
      - /path/to/your/my_shelfstone_library/config:/config # <-- UPDATE THIS
      - libreplex_cache:/cache
    ```
    Replace `/path/to/your/` with the actual absolute path to your `my_shelfstone_library` directory. Make sure these are **absolute paths**.

    **Permissions for the Configuration Directory:**
    The application inside the Docker container runs as a non-root user (default UID/GID `1001` as defined in the `Dockerfile`). This user needs read and write access to the host directory mapped to `/config` (i.e., your `my_shelfstone_library/config` directory).

    To set the correct ownership, you can use the following command, replacing `/path/to/your/my_shelfstone_library/config` with your actual path:
    ```bash
    # Determine the UID/GID the container will run as (it's 1001 in the provided Dockerfile)
    # Then, set ownership for your config directory:
    sudo chown -R 1001:1001 /path/to/your/my_shelfstone_library/config
    sudo chmod -R u+rwx /path/to/your/my_shelfstone_library/config
    ```
    If you've changed the `USER` UID/GID in the `Dockerfile`, adjust the `1001:1001` accordingly.
    Alternatively, you can grant wider permissions, but this is less secure:
    ```bash
    # Less secure: allows any user in the same group as the directory owner to write
    # sudo chmod -R g+w /path/to/your/my_shelfstone_library/config
    # Even less secure: allows any user on the system to write
    # sudo chmod -R 777 /path/to/your/my_shelfstone_library/config
    ```
    The `books` directory generally only needs to be readable by the container user, but write access might be needed if you plan to allow file deletions or modifications via ShelfStone in the future. For now, read access for the user `1001` should suffice.

4.  **Build and Run ShelfStone:**
    In the project root directory (where `docker-compose.yaml` is located), run:
    ```bash
    docker-compose up --build -d
    ```
    *   `--build`: Builds the Docker image if it's the first time or if there are changes to the `Dockerfile` or application code.
    *   `-d`: Runs the container in detached mode (in the background).

5.  **Access ShelfStone:**
    Once the container is running, the backend API will be accessible. You can check the health endpoint:
    [http://localhost:8080/api/health](http://localhost:8080/api/health)

    *(Note: The web interface and full API functionality will be developed in future phases.)*

**Stopping ShelfStone:**
```bash
docker-compose down
``` 

## Future Plans (FUTURE_PLANS.md)

ShelfStone is an evolving project. Exciting features are planned for future phases, including:

    Phase 2: MVP Features
        Automated library scanner to detect and process new ebooks.
        Core API endpoints for listing books, authors, series, serving covers, and downloading/reading books.
    Phase 3: The User Experience
        A SvelteKit frontend providing a modern, responsive UI.
        An in-browser EPUB reader (using Epub.js).
    Phase 4: Post-MVP "Plex-like" Features
        User management and authentication.
        Reading progress synchronization.
        Metadata editing via the UI.
        Advanced full-text search capabilities.

Refer to the FUTURE_PLANS.md file in this repository for more details.
Contributing

(Details about contributing will be added as the project matures.)

## Credits and Open Source

ShelfStone is built with Go and leverages several excellent open-source projects:

*   **[Calibre](https://calibre-ebook.com/):** The core ebook processing power. ShelfStone uses Calibre's command-line tools (`ebook-meta`, `ebook-convert`) for metadata handling and format conversion. Calibre is licensed under the [GNU GPL v3](https://www.gnu.org/licenses/gpl-3.0.html).
*   **[Gin Web Framework](https://gin-gonic.com/):** A high-performance HTTP web framework for Go. (MIT License)
*   **[GORM](https://gorm.io/):** The fantastic ORM library for Go, used for database interactions. (MIT License)
*   **[go-sqlite3](https://github.com/mattn/go-sqlite3):** SQLite driver for Go, enabling simple local database storage. (MIT License)
*   **[Docker](https://www.docker.com/):** For containerization and simplified deployment.

We are grateful to the developers and communities behind these projects.

## License

ShelfStone itself is planned to be released under an open-source license like MIT or Apache 2.0. The exact license is yet to be finalized.

Please note that while ShelfStone aims for a permissive license, the underlying Calibre components it utilizes are under the GNU GPL v3.

This README reflects the project's state after Phase 1. As new features are implemented, this document will be updated.