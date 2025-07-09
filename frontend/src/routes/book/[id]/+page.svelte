<script lang="ts">
  import { page } from '$app/stores';
  import { library, type Book } from '$lib/stores/library';
  import { goto } from '$app/navigation';
  import Button from '$lib/components/global/Button.svelte';

  let currentBook: Book | undefined;
  let bookId: string;

  $: bookId = $page.params.id;

  $: {
    if (bookId) {
      const foundBook = $library.find(b => b.id === bookId);
      currentBook = foundBook;
    }
  }

  function startReading() {
    if (currentBook) {
      goto(`/read?url=${encodeURIComponent(currentBook.epubUrl)}`);
    }
  }
</script>

<div class="container mx-auto py-8 px-4">
  {#if currentBook}
    <div class="bg-white shadow-xl rounded-lg p-8">
      <div class="flex flex-col md:flex-row gap-8">
        {#if currentBook.coverUrl}
          <img src={currentBook.coverUrl} alt="Cover for {currentBook.title}" class="w-full md:w-1/3 h-auto object-contain rounded-md shadow-md"/>
        {:else}
          <div class="w-full md:w-1/3 h-96 bg-gray-200 flex items-center justify-center rounded-md shadow-md">
            <span class="text-gray-500 text-xl">No Cover Available</span>
          </div>
        {/if}
        <div class="flex-1">
          <h1 class="text-4xl font-bold text-gray-800 mb-2">{currentBook.title}</h1>
          <p class="text-2xl text-gray-600 mb-6">by {currentBook.author}</p>

          <div class="mb-6">
            <h2 class="text-xl font-semibold text-gray-700 mb-2">Progress</h2>
            {#if currentBook.progress !== undefined && currentBook.progress > 0}
              <div class="w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                <div class="bg-blue-600 h-2.5 rounded-full" style="width: {currentBook.progress}%"></div>
              </div>
              <p class="text-sm text-gray-500 mt-1">{currentBook.progress}% complete</p>
            {:else}
              <p class="text-sm text-gray-500">Not started yet.</p>
            {/if}
          </div>

          <div class="mt-8">
            <Button variant="primary" onClick={startReading}>
              Read Now
            </Button>
          </div>

          <div class="mt-8 prose max-w-none">
            <h2 class="text-xl font-semibold text-gray-700 mb-2">Description</h2>
            <p class="text-gray-700">
              (Placeholder for book description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.
              Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,
              quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.)
            </p>
          </div>
        </div>
      </div>
    </div>
  {:else if bookId}
    <p class="text-center text-xl text-red-500 p-8">Book with ID '{bookId}' not found.</p>
  {:else}
    <p class="text-center text-xl text-gray-500 p-8">Loading book details...</p>
  {/if}
</div>
