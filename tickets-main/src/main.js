import Vue from 'vue'
import App from './App.vue'
import router from './router'
import BootstrapVue from 'bootstrap-vue'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import vSelect from 'vue-select'

import 'vue-select/dist/vue-select.css';

Vue.component('v-select', vSelect)

Vue.use(BootstrapVue)

Vue.config.productionTip = false

export const eventEmitter = new Vue()

new Vue({
  router,
  render: h => h(App)
}).$mount('#app')
