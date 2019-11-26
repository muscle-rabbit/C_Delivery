import Vue from "vue";
import Router from "vue-router";
import Chat from "@/components/Chat";
import OrderList from "@/components/WorkerPanel/OrderList";
import Home from "@/components/Home";
import NotFound from "@/components/NotFound";

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: "/",
      name: "Home",
      component: Home
    },
    {
      path: "/user/:userID/worker_panel",
      name: "OrderList",
      component: OrderList,
      props: true
    },
    {
      path: "/user/:userID/order/:orderID/chats/:chatID",
      name: "Chat",
      component: Chat,
      props: true
    },
    {
      path: "*",
      component: NotFound
    }
  ]
});
