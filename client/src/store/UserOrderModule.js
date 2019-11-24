import axios from "axios";

const OrderModule = {
  state: {
    order: {}
  },
  mutations: {
    FETCH_ORDER(state, payload) {
      state.order = payload;
    },
    OPEN_ORDER_DETAIL(state) {
      state.isOpen = true;
    },
    CLOSE_ORDER_DETAIL(state) {
      state.isOpen = false;
    }
  },
  actions: {
    fetchOrder({ commit }, order_id) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      axios
        .get(`/order/${order_id}`)
        .then(response => {
          commit("FETCH_ORDER", response.data);
        })
        .catch(error => {
          // eslint-disable-next-line no-console
          console.error(error.statusText);
        });
    },
    openOrderDetail({ commit }) {
      commit("OPEN_ORDER_DETAIL");
    },
    closeOrderDetail({ commit }) {
      commit("CLOSE_ORDER_DETAIL");
    }
  }
};

export default OrderModule;
