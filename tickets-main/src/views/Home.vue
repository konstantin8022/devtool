<template>
  <main>
    <Carousel></Carousel>
    <Films></Films>
    <Info></Info>
    <Footer></Footer>
  </main>
</template>

<script>
import Carousel from "@/components/Carousel";
import Films from "@/components/Films";
import Info from "@/components/Info";
import Footer from "@/components/Footer";

export default {
  name: "home",
  components: {
    Carousel,
    Films,
    Info,
    Footer
  }
};
</script>
