<script lang="ts">
  import Reader from "$lib/components/reader/Reader.svelte";
  import { page } from '$app/stores';

  let epubUrl: string | null = null;

  $: {
    const urlParam = $page.url.searchParams.get('url');
    if (urlParam) {
      epubUrl = decodeURIComponent(urlParam);
    } else {
      epubUrl = null;
    }
  }
</script>

<div class="w-full h-full" style="height: calc(100vh - var(--navbar-height, 4rem));">
  {#if epubUrl}
    <Reader {epubUrl} />
  {:else}
    <p class="text-center text-red-500 p-8">Error: No EPUB URL specified in the query parameters. Please use a link like /read?url=your_book.epub</p>
  {/if}
</div>

<style>
  /* Approximate navbar height for calc() */
  :root {
    --navbar-height: 4rem; /* Adjust if Navbar height changes */
  }
</style>
