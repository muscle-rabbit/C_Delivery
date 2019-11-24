import axios from "axios";

const WorkerPanelModule = {
  state: {
    orders: [],
    selectedOrderDoc: {}
  },
  mutations: {
    FETCH_ORDERS(state, payload) {
      state.orders = payload;
    },
    UPDATE_ORDER(state, payload) {
      state.orders.forEach((order, i) => {
        if (order.document_id === payload.document_id) {
          state.orders[i] = payload;
        }
      });
    },
    SELECT_ORDER(state, payloard) {
      state.selectedOrderDoc = payloard;
    }
  },
  actions: {
    async fetchOrders({ commit }) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      await axios
        .get("/order_list")
        .then(response => {
          commit("FETCH_ORDERS", response.data);
          // eslint-disable-next-line no-console
          console.log("fetched!");
        })
        .catch(error => {
          // eslint-disable-next-line no-console
          console.error(error.statusText);
        });
    },
    async finishTrade({ commit }, orderID) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      await axios
        .get(`/order/${orderID}/finish`)
        .then(response => {
          commit("UPDATE_ORDER", response.data);
        })
        .catch(e => {
          // eslint-disable-next-line no-console
          console.error(e);
        });
    },
    async unfinishTrade({ commit }, orderID) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      await axios
        .get(`/order/${orderID}/unfinish`)
        .then(response => {
          // eslint-disable-next-line no-console
          commit("UPDATE_ORDER", response.data);
        })
        .catch(e => {
          // eslint-disable-next-line no-console
          console.error(e);
        });
    },
    selectOrder({ commit }, orderDoc) {
      commit("SELECT_ORDER", orderDoc);
    }
  }
};

export default WorkerPanelModule;
