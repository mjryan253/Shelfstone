# ShelfStone: Your Personal Ebook Server

ShelfStone is a self-hosted ebook server designed for ease of setup and a clean, modern user experience. It aims to be a lightweight, Docker-first alternative to more complex solutions, allowing you to effortlessly manage and access your ebook library.

**Current Status: Phase 1 Complete (Foundation)**

The core backend services, database setup, and Docker containerization are complete. ShelfStone is ready for further development of its library management and user interface features.

## Core Principles

*   **Ease of Setup:** Get up and running quickly with Docker.
*   **Clean User Experience:** A modern web interface for reading and management (Future Phase).
*   **Format Handling:** Leverages the power of Calibre's command-line tools for robust ebook format support.
*   **Performance:** Built with Go (Gin) for a fast and efficient backend.
*   **Simplicity:** SQLite database for minimal external dependencies.

## Features (Phase 1 - Foundation)

*   **Go Backend (Gin Framework):** A lightweight and performant API foundation.
    *   Includes a `/api/health` endpoint for monitoring.
*   **SQLite Database (GORM):** Simple, file-based database for storing book metadata, authors, series, and user information.
    *   Automatic schema migration on startup.
*   **Calibre CLI Integration:** Wrapper functions in Go to utilize `ebook-meta` (for metadata extraction) and `ebook-convert` (for format conversion). This allows ShelfStone to handle a wide variety of ebook formats without reinventing the wheel.
*   **Dockerized Deployment:**
    *   **Multi-stage `Dockerfile`:** Creates a minimal, efficient production image containing the Go application and Calibre's CLI tools.
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
    Replace `/path/to/your/` with the actual absolute path to your `my_shelfstone_library` directory.

    **Permissions:** Ensure that the user running Docker has write permissions to your chosen host directories, especially the configuration directory. The application inside the container runs as a non-root user (`appuser`, typically UID/GID 1000 or 1001).
    You might need to adjust permissions on your host:
    ```bash
    # Example: If your user is ID 1000
    sudo chown -R 1000:1000 /path/to/your/my_shelfstone_library/config
    sudo chmod -R u+rw /path/to/your/my_shelfstone_library/config
    ```

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
License

(To be determined - likely an open-source license like MIT or Apache 2.0.)

This README reflects the project's state after Phase 1. As new features are implemented, this document will be updated.