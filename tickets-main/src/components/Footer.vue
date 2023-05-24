<template>
  <footer class="py-5 bg-light">
    <div class="container">
      <p class="m-0 text-center">Все права защищены © Слёрм ТИКЕТС {{ currentYear }}</p>
    </div>
  </footer>
</template>

<script>
export default {
  computed: {
    currentYear() {
      return new Date().getFullYear();
    }
  }
}
</script>