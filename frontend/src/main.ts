import { createApp } from "vue";
import { createPinia } from "pinia";
import App from "./App.vue";
import router from "./router";
import "vue-sonner/style.css";
import "./assets/index.css";

const app = createApp(App);

app.use(createPinia());
app.use(router);

router.isReady().then(() => {
  app.mount("#app");
});
