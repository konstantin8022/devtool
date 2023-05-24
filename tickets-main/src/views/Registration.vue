<template>
  <div class="registration">
    
    <form class="form-signin">
      <div class="text-center mb-4">
        <h1 class="h3 mb-3 font-weight-normal">Регистрация</h1>
      </div>

      <div class="form-label-group mb-3">
        <input type="email" id="inputEmail" class="form-control" placeholder="Email" required="true" autofocus="">
      </div>

      <div class="form-label-group mb-3">
        <input type="password" id="inputPassword" class="form-control" placeholder="Пароль" required="true">
      </div>

      <div class="form-label-group mb-3">
        <input type="password" id="confirmPassword" class="form-control" placeholder="Подтвердить пароль" required="true">
      </div>

      <!-- <div class="checkbox mb-3">
        <label>
          <input type="checkbox" value="remember-me"> Запомнить
        </label>
      </div> -->
      <button class="btn btn-lg btn-primary btn-block" type="submit">Подтвердить</button>
    </form>
  </div>
</template>

<style scoped>
  .registration {
    max-width: 420px;
    margin: 100px auto;
    padding: 0 20px;
  }
</style>
