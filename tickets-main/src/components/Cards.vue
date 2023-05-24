<template>
  <div class="row text-center">
    <div class="col-md-3 col-sm-6 mb-4" v-for="(movie, index) in movies" :key="movie.id">
      <div class="card h-100 border-light">

        <svg height="150" width="100%">
          <defs>
            <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="10%">
              <stop offset="0%" style="stop-color:#6c757d;stop-opacity:1" />
              <stop offset="90%" style="stop-color:#212529;stop-opacity:1" />
            </linearGradient>
          </defs>
          <rect width="100%" height="100%" fill="url(#grad1)" />
          <text x="-65%" y="140%" font-size="220" font-weight="bold" fill="#4f565c" transform="rotate(-45)">{{ movie.name }}</text>
          Sorry, your browser does not support inline SVG.
        </svg>

        <div class="card-body">
          <div>
            <span class="rating-star" v-for="n of 4">&#9733;</span>
            <span>&#9733;</span>
          </div>
          <p class="card-title">{{ movie.name }}</p>
        </div>
        <div class="card-footer">
          <button
            class="btn btn-primary"
            :id="'movie-' + index"
            @click="chooseMovie(movie), showModalFunc()"
          >
            Выберите фильм
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";
import { eventEmitter } from "@/main";

export default {
  name: "cards",

  props: {
    cityId: String,
    movies: Array
  },

  data() {
    return {
      showModal: false
    };
  },

  methods: {
    showModalFunc() {
      eventEmitter.$emit("changeModalParam", true);
      this.showModal = true;
    },

    chooseMovie(movie) {
      this.movieTitle = movie.title;
      this.$emit("knowMovieTitle", this.movieTitle);
      this.movieId = movie.id;
      this.$emit("knowMovieId", this.movieId);

      axios
        .get(
          `${process.env.VUE_APP_API_URL}/cities/${this.cityId}/movies/${this.movieId}/seances`
        )
        .then(response => {
          console.log(response.data.data);
          this.seances = response.data.data;
          this.$emit("knowSeances", this.seances);
        })
        .catch(error => {
          console.log("-----error-------");
          console.log(error);
        });
    },

    makeToast(text, variant = null, append = false) {
      this.toastCount++;
      this.$bvToast.toast(text, {
        title: "Оповещение от сервиса Tickets",
        variant: variant,
        autoHideDelay: 3000,
        appendToast: append
      });
    }
  }
};
</script>

<style scoped>
.card {
  transition: 0.5s;
}

.card:hover {
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.1);
}

.card-footer {
  background-color: white;
  border-top: none;
}

.rating-star {
  color: #007bff;
}
</style>