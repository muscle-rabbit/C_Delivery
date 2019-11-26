import axios from "axios";

const UserModule = {
  state: {
    user: {}
  },
  mutations: {
    FETCH_USER(state, payload) {
      state.user = payload;
    }
  },
  actions: {
    fetchUser({ commit }, id) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      axios
        .get(`/api/v1/user/${id}`)
        .then(response => {
          commit("FETCH_USER", response.data);
        })
        .catch(error => {
          // eslint-disable-next-line no-console
          console.error(error.statusText);
        });
    }
  }
};

export default UserModule;
