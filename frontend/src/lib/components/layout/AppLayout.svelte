<script lang="ts">
  import Navbar from "./Navbar.svelte";
  import Sidebar from "./Sidebar.svelte";
</script>

<div class="flex flex-col h-screen">
  <Navbar />
  <div class="flex flex-1 overflow-hidden">
    <Sidebar />
    <main class="flex-1 p-6 overflow-y-auto bg-gray-100">
      <slot />
    </main>
  </div>
</div>
