<template>
  <section class="py-5 bg-light">
    <div class="container">

      <div class="d-flex justify-content-between mb-3">
        <h3 class="mt-2">Фильмы</h3>
        <div class="invoice-item_course" style="width:200px;">
          <v-select
            :multiple="false"
            :options="cities"
            label="name"
            v-model="cityName"
            :reduce="(city) => city.name"
          />
        </div>
      </div>

      <Cards
        :movies="movies"
        :cityId="cityId"
        @knowMovieTitle="movieTitle = $event"
        @knowMovieId="movieId = $event"
        @knowSeances="seances = $event"
      >
      </Cards>

      <Modal
        :movieId="movieId"
        :movieTitle="movieTitle"
        :cityId="cityId"
        :cityName="cityName"
        :seances="seances"
      >
      </Modal>

      <div class="text-center mt-3">
        <button
          class="btn btn-secondary"
          @click="increaseFilmsCounter"
        >
          <img
            src="../assets/refresh.svg"
            alt="refresh"
          />
          Загрузить ещё
        </button>
      </div>

    </div>

  </section>
</template>

<script>
import Cards from '@/components/Cards.vue'
import Modal from '@/components/Modal.vue'
import axios from 'axios'

export default {
  name: 'films',

  components: {
    Modal,
    Cards,
  },

  data() {
    return {
      cities: [],
      cityId: null,
      cityName: 'Город',
      movies: [],
      movieId: null,
      movieTitle: null,
      seances: null,
      filmsCounter: 8
    }
  },

  watch: {
    cityName() {
      if (this.cityName) {
        let city = this.cities.find(c => c.name === this.cityName);
        this.getAllMoviesInCity(city.id);
      }
    }
  },

  created() {
    this.getAllCities();
  },

  methods: {
    increaseFilmsCounter() {
      this.filmsCounter += 8;
      let city = this.cities.find(c => c.name === this.cityName);
      this.getAllMoviesInCity(city.id);
    },

    getAllCities() {
      axios.get(`${process.env.VUE_APP_API_URL}/cities`)
        .then(response => {
          this.cities = response.data.data;
          this.cityId = response.data.data[0].id
          this.cityName = response.data.data[0].name
          console.log(response.data.data)
        })
        .catch(error => {
          console.log('-----error-------');
          console.log(error);
          this.makeToast('Ошибка', "danger");
        })
    },

    getAllMoviesInCity(cityId) {
      // console.log(cityId)
      axios.get(`${process.env.VUE_APP_API_URL}/cities/${cityId}/movies`, {
        params: { max_results: this.filmsCounter, with_seances: true }
      })
        .then(response => {
          this.movies = response.data.data;
          console.log(response.data.data)
        })
        .catch(error => {
          console.log('-----error-------');
          console.log(error);
          this.makeToast('Ошибка', "danger");
        })
    },

    makeToast(text, variant = null, append = false) {
      this.toastCount++;
      this.$bvToast.toast(text, {
        title: "Оповещение от сервиса Tickets",
        variant: variant,
        autoHideDelay: 3000,
        appendToast: append
      });
    },
  }
}
</script>