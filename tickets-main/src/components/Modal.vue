<template>
  <div>
    <transition name="fade">
      <div class="modal fade show" tabindex="-1" role="dialog" v-if="showModal">
        <div class="modal-dialog modal-lg" role="document">
          <div class="modal-content">
            <div class="modal-header">
              <h3 class="modal-title" id="movie-title">{{ movieTitle }}</h3>
              <button
                type="button"
                class="close"
                data-dismiss="modal"
                aria-label="Close"
                @click="closeModal"
              >
                <span aria-hidden="true">&times;</span>
              </button>
            </div>
            <div class="modal-body">
              <div class="d-flex justify-content-start mb-3">
                <b-dropdown left text="Выберите время" class="m-2" id="seances-dropdown">
                  <b-dropdown-item-button
                    v-for="(seance, index) in seances"
                    :key="seance.id"
                    @click="chooseSeance(seance)"
                    :id="'seance-' + index"
                  >{{ formatDateTime(seance.datetime) }}</b-dropdown-item-button>
                </b-dropdown>
              </div>

              <transition name="fade" mode="out-in">
                <!-- cinema hall -->
                <div v-if="seanceId" key="cinema-hall" id="cinema-hall">
                  <hr class="screen mt-4" />
                  <table class="table table-borderless mb-3">
                    <tbody>
                      <tr v-for="row in 5">
                        <th scope="row">{{ row }}</th>
                        <td v-for="i of 10">
                          <span class="not-available">{{ i }}</span>
                        </td>
                      </tr>
                      <tr>
                        <th scope="row">6</th>
                        <td v-for="(seat, index) of seanceSeats" :key="seat.id">
                          <span
                            :class="seat.vacant ? 'seat' : 'not-available'"
                            @click="addSeatId(seat, $event)"
                            :id="'seat-' + index"
                          >{{ seat.id }}</span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <div class="lds-grid" v-else key="spinner">
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                  <div></div>
                </div>
              </transition>

              <div class="email-input card p-3 mt-5">
                <p id="details">Время: {{ formatSeanceDate }}</p>
                <p id="sum">Cумма: {{ calcSum }} руб.</p>
                <input
                  type="email"
                  class="form-control"
                  :class="{'email-error': emailError}"
                  v-model.lazy="email"
                  @click="emailError = false"
                  placeholder="Введите Email"
                  id="email-field"
                />

                <input
                  type="card"
                  class="form-control"
                  :class="{'card-error': cardError}"
                  v-model.lazy="card"
                  @click="cardError = false"
                  placeholder="Введите номер карты"
                  id="card-field"
                />
              </div>
            </div>

            <div class="modal-footer">
              <button
                type="button"
                class="btn btn-outline-primary"
                data-dismiss="modal"
                @click="closeModal"
                id="close-modal"
              >Закрыть</button>
              <button
                type="button"
                class="btn btn-outline-primary"
                @click="cleanForm"
                id="clean-form"
              >Очистить</button>
              <button
                type="submit"
                class="btn btn-primary"
                @click="checkEmailAndSeats"
                id="submit-button"
              >Отправить</button>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <transition name="fade">
      <div class="modal-backdrop fade show" v-if="showModal"></div>
    </transition>
  </div>
</template>

<script>
import axios from "axios";
import { eventEmitter } from "@/main";

export default {
  name: "modal",

  props: {
    movieTitle: String,
    movieId: Number,
    cityId: String,
    cityName: String,
    seances: Array
  },

  data() {
    return {
      moviePrice: null,
      seancePrice: null,
      seanceId: null,
      seanceDate: "",
      seanceSeats: null,
      email: null,
      card: null,
      seatsIds: [],
      emailError: false,
      cardError: false,
      showModal: false,
      toastCount: 0
    };
  },

  computed: {
    calcSum() {
      return this.seatsIds.length * this.seancePrice;
    },
    formatSeanceDate() {
      if (this.seanceDate) {
        let date = this.seanceDate;
        return `${date.slice(8, 10)} ${this.formatMonth(date)} в ${date.slice(
          11,
          16
        )}`;
      } else {
        return "";
      }
    }
  },

  created() {
    eventEmitter.$on("changeModalParam", val => {
      this.showModal = val;
    });
  },

  methods: {
    formatMonth(date) {
      let month;
      if (date.slice(5, 7) == "01") {
        month = "января";
      } else if (date.slice(5, 7) == "02") {
        month = "февраля";
      } else if (date.slice(5, 7) == "03") {
        month = "марта";
      } else if (date.slice(5, 7) == "04") {
        month = "апреля";
      } else if (date.slice(5, 7) == "05") {
        month = "мая";
      } else if (date.slice(5, 7) == "06") {
        month = "июня";
      } else if (date.slice(5, 7) == "07") {
        month = "июля";
      } else if (date.slice(5, 7) == "08") {
        month = "августа";
      } else if (date.slice(5, 7) == "09") {
        month = "сентября";
      } else if (date.slice(5, 7) == "10") {
        month = "октября";
      } else if (date.slice(5, 7) == "11") {
        month = "ноября";
      } else if (date.slice(5, 7) == "12") {
        month = "декабря";
      } else {
        month = "ошибка";
      }
      return month;
    },
    formatDateTime(date) {
      return `${date.slice(8, 10)} ${this.formatMonth(date)} в ${date.slice(
        11,
        16
      )}`;
    },

    forceHallRerender(seance) {
      this.seanceId = false;
      this.$nextTick(() => {
        this.seanceId = seance.id;
      });
    },

    chooseSeance(seance) {
      this.forceHallRerender(seance);
      this.seatsIds = [];
      this.seancePrice = seance.price;
      this.seanceSeats = seance.seats;
      this.seanceDate = seance.datetime;
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

    closeModal() {
      this.cleanForm();
      this.showModal = false;
      this.makeToast("Ваш заказ был отменён", "primary");
    },

    checkEmailAndSeats() {
      let validEmail = /\S+@\S+\.\S+/.test(this.email);

      if (validEmail && this.seatsIds.length > 0 && this.card != null && this.card.length > 0) {
        this.formSubmit();
      } else {
        this.emailError = true;
        this.makeToast("Информация о заказе заполнена некорректно", "danger");
      }
    },

    formSubmit() {
      console.log(`${process.env.VUE_APP_API_URL}/cities/${this.cityId}/movies/${this.movieId}/seances/${this.seanceId}/bookings
                    email: ${this.email}
                    seats: ${this.seatsIds}`);

      axios
        .post(
          `${process.env.VUE_APP_API_URL}/cities/${this.cityId}/movies/${this.movieId}/seances/${this.seanceId}/bookings`,
          {
            email: this.email,
            seatsIds: this.seatsIds,
            card: this.card
          }
        )
        .then(response => {
          console.log(response);
          this.makeToast("Заказ был успешно выполнен", "success");
        })
        .catch(error => {
          console.log("-----error-------");
          console.log(error);
          this.makeToast("Ошибка оформления заказа", "danger");
        });
      this.cleanForm();
      this.showModal = false;
    },
    addSeatId(seat, event) {
      if (this.seatsIds.indexOf(seat.id) === -1 && seat.vacant) {
        this.seatsIds.push(seat.id);
        event.target.style.backgroundColor = "#007bff";
        event.target.style.color = "white";
      }
    },
    cleanForm() {
      this.seanceId = null;
      (this.seanceDate = ""), (this.email = null);
      this.seatsIds = [];
    }
  }
};
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.5s;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}

.email-error {
  border: 1px solid rgb(255, 174, 174);
  box-shadow: 0 0 0 0.2rem rgba(255, 174, 174, 0.5);
}

.modal {
  display: block;
}

.seat {
  cursor: pointer;
  display: inline-block;
  color: #007bff;
  width: 25px;
  height: 25px;
  vertical-align: middle;
  margin: 0 5px;
  line-height: 1.5rem;
  border: 1px solid #007bff;
  border-radius: 50%;
  text-align: center;
  transition: 0.5s;
}

.seat:hover {
  background-color: #007bff;
  color: white;
}

.email-input {
  padding: 0 16px;
}

.not-available {
  cursor: not-allowed;
  display: inline-block;
  color: #999999;
  width: 25px;
  height: 25px;
  line-height: 1.5rem;
  margin: 0 5px;
  border: 1px solid #999999;
  border-radius: 50%;
  text-align: center;
}

table {
  width: auto;
  margin: 0 auto;
  font-size: 12px;
}

td {
  padding: 5px 0 !important;
  margin: 5px !important;
}

th {
  padding: 5px 20px !important;
  font-weight: 200;
  color: #aaa;
}

.screen {
  height: 3px;
  border: none;
  color: #0069d9;
  background-color: #0069d9;
  margin-bottom: 40px;
  width: 400px;
}

/* spinner */
.lds-grid {
  display: block;
  position: relative;
  width: 80px;
  height: 80px;
  margin: 110px auto 135px;
}
.lds-grid div {
  position: absolute;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #007bff;
  animation: lds-grid 1.2s linear infinite;
}
.lds-grid div:nth-child(1) {
  top: 8px;
  left: 8px;
  animation-delay: 0s;
}
.lds-grid div:nth-child(2) {
  top: 8px;
  left: 32px;
  animation-delay: -0.4s;
}
.lds-grid div:nth-child(3) {
  top: 8px;
  left: 56px;
  animation-delay: -0.8s;
}
.lds-grid div:nth-child(4) {
  top: 32px;
  left: 8px;
  animation-delay: -0.4s;
}
.lds-grid div:nth-child(5) {
  top: 32px;
  left: 32px;
  animation-delay: -0.8s;
}
.lds-grid div:nth-child(6) {
  top: 32px;
  left: 56px;
  animation-delay: -1.2s;
}
.lds-grid div:nth-child(7) {
  top: 56px;
  left: 8px;
  animation-delay: -0.8s;
}
.lds-grid div:nth-child(8) {
  top: 56px;
  left: 32px;
  animation-delay: -1.2s;
}
.lds-grid div:nth-child(9) {
  top: 56px;
  left: 56px;
  animation-delay: -1.6s;
}
@keyframes lds-grid {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>