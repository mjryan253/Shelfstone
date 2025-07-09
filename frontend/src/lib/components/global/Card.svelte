<script lang="ts">
  export let title: string = "";
  export let coverUrl: string | null = null;
  // Placeholder for progress, e.g., percentage or pages read
  export let progressText: string | null = null;
</script>

<div class="bg-white shadow-lg rounded-lg overflow-hidden">
  {#if coverUrl}
    <img src={coverUrl} alt="Cover for {title}" class="w-full h-48 object-cover" />
  {:else}
    <div class="w-full h-48 bg-gray-200 flex items-center justify-center">
      <span class="text-gray-500">No Cover</span>
    </div>
  {/if}
  <div class="p-4">
    <h3 class="font-bold text-lg mb-2 truncate" title={title}>{title}</h3>
    {#if progressText}
      <p class="text-gray-600 text-sm">Progress: {progressText}</p>
    {/if}
    <slot name="actions"></slot> <!-- For buttons like "Read Now", "Details" -->
  </div>
</div>
