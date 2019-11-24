import axios from "axios";

const ChatModule = {
  state: {
    chats: {}
  },
  getters: {
    getChats: state => {
      return state.chats;
    }
  },
  mutations: {
    FETCH_CHATS(state, payload) {
      state.chats = payload;
    }
  },
  actions: {
    async fetchChats({ commit }, id) {
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      await axios
        .get(`/chats/${id}`)
        .then(response => {
          commit("FETCH_CHATS", response.data);
        })
        .catch(error => {
          // eslint-disable-next-line no-console
          console.error(error.statusText);
        });
    },
    // eslint-disable-next-line no-unused-vars
    sendMessage({ dispatch, commit }, payload) {
      let chatID = payload.chatID;
      const message = {
        user_id: payload.user_id,
        content: payload.content
      };
      axios.defaults.baseURL = process.env.VUE_APP_API_ENDPOINT;
      axios
        .post(`/chats/${chatID}`, message)
        // eslint-disable-next-line no-unused-vars
        .then(r => {
          dispatch("fetchChats", chatID);
        })
        .catch(e => {
          // eslint-disable-next-line no-console
          console.error(e);
        });
    }
  }
};

export default ChatModule;
