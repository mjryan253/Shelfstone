<script lang="ts">
  export let type: "button" | "submit" | "reset" = "button";
  export let variant: "primary" | "secondary" | "danger" = "primary";
  export let disabled: boolean = false;
  export let onClick: (() => void) | null = null;

  let baseClasses = "font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline";
  let variantClasses = "";

  $: {
    switch (variant) {
      case "primary":
        variantClasses = "bg-blue-500 hover:bg-blue-700 text-white";
        break;
      case "secondary":
        variantClasses = "bg-gray-500 hover:bg-gray-700 text-white";
        break;
      case "danger":
        variantClasses = "bg-red-500 hover:bg-red-700 text-white";
        break;
    }
    if (disabled) {
      variantClasses += " opacity-50 cursor-not-allowed";
    }
  }
</script>

<button {type} class="{baseClasses} {variantClasses}" {disabled} on:click={onClick}>
  <slot />
</button>
