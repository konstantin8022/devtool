<template>
  <b-navbar toggleable="sm">
    <b-container>

      <b-navbar-brand to="/"><img alt="Vue logo" src="../assets/logo.png" width="80"></b-navbar-brand>

      <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>

      <b-collapse id="nav-collapse" is-nav>
        <b-navbar-nav class="ml-auto">
          <b-nav-item to="/login">Вход</b-nav-item>
          <b-nav-item to="/registration">Регистрация</b-nav-item>
        </b-navbar-nav>
      </b-collapse>

    </b-container>
  </b-navbar>
</template>

<script>
export default {

}
</script>

<style>

</style>