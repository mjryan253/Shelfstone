<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import ePub, { type Book, type Rendition } from "epubjs";

  export let epubUrl: string;

  let book: Book | null = null;
  let rendition: Rendition | null = null;
  let readerElement: HTMLDivElement;
  let currentLocationDisplay: string = "Loading...";
  let isLoading: boolean = true;
  let errorMessage: string | null = null;

  onMount(async () => {
    if (!epubUrl) {
      errorMessage = "No EPUB URL provided.";
      isLoading = false;
      return;
    }

    try {
      book = ePub(epubUrl);
      rendition = book.renderTo(readerElement, {
        width: "100%",
        height: "100%",
        flow: "paginated", // Or "scrolled-doc" for continuous scrolling
        // spread: "auto", // For two-page spreads on wider screens
      });

      await rendition.display();
      isLoading = false;

      rendition.on("displayed", () => {
        // This event might be more reliable for initial display
      });

      rendition.on("relocated", (location: any) => {
        if (location && location.start) {
          const cfi = location.start.cfi;
          if (book && book.locations) {
            const page = book.locations.locationFromCfi(cfi);
            const totalPages = book.locations.length();
            currentLocationDisplay = `Page ${page + 1} of ${totalPages}`;
          } else {
            currentLocationDisplay = `Location: ${cfi}`;
          }
        }
      });

      // Generate locations for page numbers (can be slow for large books)
      // Consider doing this in a web worker or on demand
      await book.ready; // Ensure book metadata and spine are loaded
      if (book.locations) { // Check if locations are already generated (e.g. from metadata)
        // Do nothing if already generated
      } else {
        await book.locations.generate(1024); // Generate locations based on 1024 characters per page segment
      }
      // After locations are generated, update current location display
      if (rendition && rendition.currentLocation()) {
        const cfi = rendition.currentLocation().start.cfi;
        if (book && book.locations) {
            const page = book.locations.locationFromCfi(cfi);
            const totalPages = book.locations.length();
            currentLocationDisplay = `Page ${page + 1} of ${totalPages}`;
          }
      }


    } catch (error) {
      console.error("Error loading EPUB:", error);
      errorMessage = `Error loading EPUB: ${error instanceof Error ? error.message : String(error)}`;
      isLoading = false;
    }
  });

  onDestroy(() => {
    book?.destroy();
  });

  async function nextPage() {
    await rendition?.next();
  }

  async function prevPage() {
    await rendition?.prev();
  }
</script>

<div class="w-full h-full flex flex-col bg-gray-200">
  {#if isLoading}
    <div class="flex-grow flex items-center justify-center">
      <p>Loading EPUB...</p>
      <!-- TODO: Add a spinner component -->
    </div>
  {:else if errorMessage}
    <div class="flex-grow flex items-center justify-center p-4">
      <p class="text-red-500">{errorMessage}</p>
    </div>
  {:else}
    <div bind:this={readerElement} class="flex-grow overflow-hidden bg-white shadow-inner">
      <!-- Epub.js will render here -->
    </div>
    <div class="flex justify-between items-center p-2 bg-gray-300 shadow-md">
      <button on:click={prevPage} class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50" disabled={!rendition}>
        Previous
      </button>
      <div class="text-sm text-gray-700">
        {currentLocationDisplay}
      </div>
      <button on:click={nextPage} class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50" disabled={!rendition}>
        Next
      </button>
    </div>
  {/if}
</div>

<style>
  /* Ensure reader element takes up space */
  .flex-grow {
    flex-grow: 1;
  }
 /* Basic styling for the rendition viewport to avoid issues */
  :global(.epub-container) {
    overflow: hidden;
    position: relative;
  }
  :global(.epub-view > iframe) {
    width: 100% !important; /* Ensure iframe takes full width of its container */
    height: 100% !important;/* Ensure iframe takes full height of its container */
    border: none;
  }
</style>
