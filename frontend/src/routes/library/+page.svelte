<script lang="ts">
  import { library, type Book } from '$lib/stores/library';
  import Card from '$lib/components/global/Card.svelte';
  import Button from '$lib/components/global/Button.svelte';
  import { goto } from '$app/navigation';

  let books: Book[] = [];
  library.subscribe(value => {
    books = value;
  });

  function viewBookDetails(bookId: string) {
    goto(`/book/${bookId}`);
  }

  function readBook(epubUrl: string, event: MouseEvent) {
    event.stopPropagation(); // Prevent card click if button is on card
    // For now, we use a generic read page.
    // Later, this might pass the book ID or a more specific URL.
    goto(`/read?url=${encodeURIComponent(epubUrl)}`);
  }
</script>

<div class="container mx-auto py-8">
  <h1 class="text-3xl font-bold mb-8 text-gray-800">My Library</h1>

  {#if books.length === 0}
    <p class="text-gray-600">Your library is empty. Add some books!</p>
  {:else}
    <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-6">
      {#each books as book (book.id)}
        <div on:click={() => viewBookDetails(book.id)} class="cursor-pointer">
          <Card title={book.title} coverUrl={book.coverUrl} progressText={book.progress !== undefined ? `${book.progress}%` : 'Not started'}>
            <svelte:fragment slot="actions">
              <div class="mt-4 flex justify-end">
                <Button variant="primary" onClick={(event) => readBook(book.epubUrl, event)}>
                  Read
                </Button>
              </div>
            </svelte:fragment>
          </Card>
        </div>
      {/each}
    </div>
  {/if}
</div>
