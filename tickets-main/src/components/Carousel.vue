<template>
  <div class="carousel bg-light">
    <b-container>

      <b-carousel
        id="carousel-fade"
        style="text-shadow: 0px 0px 2px #000"
        fade
        indicators
        img-width="1024"
        img-height="480"
      >
        <b-carousel-slide
          caption="Звёздные Войны: Скайуокер"
          img-src="./images/carousel/film-1.jpg"
        ></b-carousel-slide>
        <b-carousel-slide
          caption="Джуманджи: Новый уровень"
          img-src="./images/carousel/film-2.jpg"
        ></b-carousel-slide>
        <b-carousel-slide
          caption="Ford против Ferrari"
          img-src="./images/carousel/film-3.jpg"
        ></b-carousel-slide>
      </b-carousel>
    </b-container>
  </div>
</template>

<style>

</style>