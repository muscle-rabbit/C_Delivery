import Vue from "vue";
import Vuex from "vuex";

import UserModule from "./UserModule";
import ChatModule from "./ChatModule";
import UserOrderModule from "./UserOrderModule";
import WorkerPanelModule from "./WorkerPanelModule";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    user: UserModule,
    chat: ChatModule,
    order: UserOrderModule,
    workerPanel: WorkerPanelModule
  },
  state: {
    loading: false,
    error: null,
    onlineUsers: []
  },
  mutations: {
    setLoading(state, payload) {
      state.loading = payload;
    },
    setError(state, payload) {
      state.error = payload;
    },
    clearError(state) {
      state.error = null;
    },
    setOnlineUsers(state, payload) {
      state.onlineUsers = payload;
    }
  },
  actions: {
    clearError({ commit }) {
      commit("clearError");
    }
  },
  getters: {
    loading(state) {
      return state.loading;
    },
    error(state) {
      return state.error;
    },
    onlineUsers(state) {
      return state.onlineUsers;
    }
  }
});
