# LibrePlex: Future Development Phases

This document outlines the planned features and development phases for LibrePlex beyond the initial foundational setup.

## Phase 2: Minimum Viable Product (MVP) Features

This is the core functionality that makes the server useful.

### 1. Library Scanner
The server will monitor a `/books` directory (configurable, maps to the Docker volume). When a new ebook file is added, it will trigger a processing pipeline.

*   **Workflow:**
    1.  **Detect:** A file watcher (e.g., using a library like `fsnotify` in Go) notices a new file in the `/books` directory (e.g., `some-book.azw3`).
    2.  **Ingest:** The file is copied to a temporary processing location (e.g., within `/cache` or a temporary system directory). Consider a queue for handling multiple additions.
    3.  **Metadata Extraction:**
        *   Run `ebook-meta` (via the Go wrapper) on the ingested file to pull metadata (title, author, series, ISBN, publisher, publication date, description, language).
        *   Extract the cover image. `ebook-meta` can export it, or it might need to be extracted from the EPUB/MOBI structure if `ebook-meta` doesn't provide a direct command. This cover should be stored in a manageable way (e.g., `/cache/covers/book_id.jpg` or alongside the book in an organized library structure).
    4.  **Database Entry:**
        *   Parse the metadata.
        *   Find or create `Author` record.
        *   Find or create `Series` record (if applicable).
        *   Create a new `Book` record, linking to the author and series, and storing all extracted metadata. Store the path to the original file and the cover image.
    5.  **Standardization (Recommended):**
        *   Run `ebook-convert` (via the Go wrapper) to create a universal `EPUB` version of the book. This EPUB will be used for the web reader.
        *   Store this standardized EPUB in a designated location (e.g., `/cache/processed_books/book_id.epub` or within the organized library structure). Update the `Book` record with the path to this EPUB.
    6.  **Organize (Optional but good for user library management):**
        *   Move and rename the original file and its cover (if extracted separately) to a clean, organized directory structure within the main `/books` volume. For example: `/books/Author Name/Series Name (if any)/Book Title (Year)/Book Title.original_ext` and `/books/Author Name/Series Name (if any)/Book Title (Year)/cover.jpg`.
        *   The paths stored in the database should reflect this organized structure.
        *   Alternatively, keep original files as is and manage processed versions and covers in `/cache` or a separate structured directory linked by the database. This might be simpler than reorganizing the user's `/books` folder directly. *Decision: Initially, it might be better to store processed files (like the web-friendly EPUB and extracted cover) in a separate application-managed directory (e.g., inside `/config/data` or `/cache/persistent_data`) rather than modifying the user's `/books` structure directly. This avoids altering user-organized files.*

### 2. Core API Endpoints
A clean RESTful API is needed for the frontend. All endpoints should be under `/api/v1/`.

*   **Authentication:** While full user management is Phase 4, consider a simple API key or basic auth for early protection if desired, though for a personal server, this might be deferred.
*   **Books:**
    *   `GET /api/books`: List all books.
        *   Support pagination (e.g., `?page=1&limit=20`).
        *   Support sorting (e.g., `?sort=title_asc`, `?sort=added_at_desc`).
        *   Support filtering (e.g., `?author_id=X`, `?series_id=Y`, `?search=query`).
        *   Response should include essential book details (ID, title, author name, series name, cover URL).
    *   `GET /api/books/{id}`: Get detailed information for a single book, including all metadata fields.
    *   `GET /api/books/{id}/cover`: Serve the book's cover image. This could be a direct file path if covers are stored predictably, or a handler that fetches the file.
    *   `GET /api/books/{id}/download?format=epub`: Download the book file.
        *   If `format` is the original format and it exists, serve it.
        *   If `format` is requested (e.g., EPUB) and a standardized version exists, serve that.
        *   (Future) If `format` is requested and doesn't exist but can be converted, trigger on-the-fly conversion via `ebook-convert` (this can be resource-intensive; consider caching results). For MVP, might only serve existing formats.
    *   `GET /api/books/{id}/read`: Serve the standardized EPUB file specifically for the web reader. This ensures the web reader always gets a compatible format.
*   **Authors:**
    *   `GET /api/authors`: List all authors with pagination and searching.
    *   `GET /api/authors/{id}`: Get details for a single author (name, list of their books).
*   **Series:**
    *   `GET /api/series`: List all series with pagination and searching.
    *   `GET /api/series/{id}`: Get details for a single series (name, list of books in the series, ordered by series index).
*   **Other (Future Considerations for MVP):**
    *   `POST /api/scan-library`: Manually trigger a full library scan.
    *   `GET /api/stats`: Basic library statistics (total books, authors, etc.).

## Phase 3: The User Experience

This is where we win against Calibre's web interface.

### 1. Frontend Technology: SvelteKit
*   **Why SvelteKit?** Modern, simple, fast, excellent developer experience. Compiles to optimized vanilla JS.
*   **Setup:**
    *   Initialize a new SvelteKit project in a separate `/frontend` directory or a separate repository.
    *   Configure it to proxy API requests to the Go backend during development (e.g., `vite.config.js` server proxy).
*   **Key Components/Views:**
    *   **Layout:** Main navigation (Library, Authors, Series, Search).
    *   **Library View (`/` or `/books`):**
        *   Grid of book covers.
        *   Infinite scrolling or pagination.
        *   Basic filtering and sorting options UI.
        *   Clicking a book navigates to its detail page.
    *   **Book Detail Page (`/books/{id}`):**
        *   Display full book metadata (cover, title, author, series, description, etc.).
        *   "Read" button (links to Reader View).
        *   "Download" button/options.
    *   **Author List Page (`/authors`):** List authors; clicking an author goes to an author detail page showing their books.
    *   **Series List Page (`/series`):** List series; clicking a series goes to a series detail page showing its books in order.
    *   **Search Bar:** Prominently displayed, allowing users to search for books, authors, and series. Search results page.

### 2. In-Browser Web Reader
*   **Solution:** Integrate **Epub.js** (or a similar modern alternative if one emerges).
*   **Implementation:**
    *   Create a "Reader" view/route in SvelteKit (e.g., `/read/{bookId}`).
    *   This view will:
        *   Fetch the book's standardized EPUB file from the backend (`GET /api/books/{id}/read`).
        *   Initialize Epub.js with the fetched EPUB.
        *   Provide UI controls for:
            *   Pagination (next/previous page/chapter).
            *   Table of Contents navigation.
            *   Font size adjustment.
            *   Theme selection (day/night mode, sepia).
            *   (Future) Bookmarking, annotations.

## Phase 4: Post-MVP "Plex-like" Features

Once the core is stable, these features will make it truly competitive.

### 1. User Management
*   **Goal:** Allow multiple users with separate libraries (optional) or at least separate reading progress and shelves.
*   **Implementation:**
    *   Expand the `User` model in Go.
    *   Implement JWT-based authentication:
        *   `POST /api/auth/register`
        *   `POST /api/auth/login` (returns JWT)
        *   `POST /api/auth/refresh-token`
    *   Protect relevant API endpoints, requiring a valid JWT.
    *   Frontend will need login/registration forms and manage JWT storage (e.g., HttpOnly cookies or local storage with care).
    *   Associate books/reading progress with users. (This might mean `UserBooks` join table for progress, ratings etc., or a `UserID` foreign key on such tables).

### 2. Reading Progress Sync
*   **Goal:** Store and sync the user's current location in a book across devices/sessions.
*   **Implementation:**
    *   **Database:** Add fields to a `UserBookProgress` table (or similar) to store `UserID`, `BookID`, `epubCfi` (Epub.js Current Function Invocation for location), `lastReadDate`, `progressPercentage`.
    *   **API:**
        *   `POST /api/books/{id}/progress`: Endpoint for the client (web reader) to periodically send the current reading location (Epub CFI string, percentage). Backend saves this to the database.
        *   `GET /api/books/{id}/progress`: Endpoint for the client to fetch the last known reading location when opening a book.
    *   **Frontend (Epub.js):**
        *   On opening a book, fetch last known location and tell Epub.js to navigate there.
        *   Listen for location changes in Epub.js and periodically send updates to the backend.

### 3. Metadata Editing
*   **Goal:** Allow users to correct or enhance book metadata directly from the web UI.
*   **Implementation:**
    *   **API:**
        *   `PUT /api/books/{id}`: Update book details.
        *   `PUT /api/authors/{id}`: Update author details.
        *   `PUT /api/series/{id}`: Update series details.
        *   (Consider how to handle cover image updates – perhaps a separate `POST /api/books/{id}/cover` endpoint).
    *   **Frontend:**
        *   Add "Edit" buttons on book detail pages, author pages, series pages.
        *   Display forms pre-filled with current metadata, allowing users to modify and save.
        *   For covers, allow uploading a new image.

### 4. Full-Text Search (Advanced)
*   **Goal:** Provide faster and more relevant search results, including searching within book contents (if feasible and desired).
*   **Solution:** Integrate a dedicated search engine like **Meilisearch** or **Typesense**.
*   **Implementation:**
    *   Add Meilisearch (or chosen alternative) as another service in `docker-compose.yaml`.
    *   When books are scanned or metadata is updated, the Go backend also sends the relevant data (title, author, series, description, ISBN, maybe even extracted text content) to Meilisearch for indexing.
    *   The API's search endpoints (`GET /api/books?search=...`, etc.) would query Meilisearch instead of directly hitting the SQLite database with `LIKE` clauses.
    *   Frontend search bar queries these API endpoints.

### 5. Additional "Plex-like" Enhancements
*   **Collections/Shelves:** Allow users to create custom collections of books.
*   **Smart Playlists/Collections:** Auto-generated collections based on criteria (e.g., "Recently Added," "Unread in X Genre").
*   **Push Notifications:** (Optional, more complex) Notify users when new books by favorite authors are added.
*   **OPDS Feed:** Provide an OPDS (Open Publication Distribution System) feed for compatibility with many third-party reader apps.
*   **Background Task Management UI:** A simple view to see the status of ongoing tasks (e.g., library scans, book conversions).
*   **Multi-format support:** More robust handling of various ebook formats beyond EPUB for reading, potentially converting to EPUB on-the-fly or pre-converting more formats.
*   **Mobile-first Frontend:** Ensure the SvelteKit UI is fully responsive and works well on mobile devices.
*   **Accessibility (a11y):** Design the frontend with accessibility best practices.
*   **Internationalization (i18n) and Localization (l10n):** Support for multiple languages in the UI.
*   **Plugin System:** (Very advanced) Allow for community plugins to extend functionality.
