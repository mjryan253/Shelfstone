<script lang="ts">
  // Placeholder for navigation links
  const navLinks = [
    { href: "/library", text: "Library" },
    { href: "/series", text: "Series" },
    { href: "/lists", text: "Reading Lists" },
    { href: "/settings", text: "Settings" },
  ];
</script>

<aside class="bg-gray-700 text-white w-64 min-h-screen p-4 space-y-2">
  <h2 class="text-lg font-semibold mb-4">Navigation</h2>
  <ul>
    {#each navLinks as link}
      <li>
        <a href={link.href} class="block py-2 px-3 rounded hover:bg-gray-600">{link.text}</a>
      </li>
    {/each}
  </ul>
</aside>
