<template>
  <div class="container">
    <v-flex style="position: relative;" class="inner">
      <div class="chat-container" id="chat-container">
        <message :ownID="ownID" :chats="chats"></message>
        <v-flex class="typer">
          <input
            type="text"
            placeholder="メッセージを入力する。"
            v-on:keyup.enter="sendMessage"
            v-model="content"
          />
          <v-icon @click="sendMessage" large color="blue darken-3"
            >mdi-send</v-icon
          >
        </v-flex>
      </div>
    </v-flex>
  </div>
</template>

<script>
import Message from "./Message";
export default {
  components: {
    Message
  },
  props: {
    chatID: {
      type: String,
      required: true
    },
    orderID: {
      type: String,
      required: true
    },
    ownID: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      content: ""
    };
  },
  computed: {
    chats() {
      return this.$store.state.chat.chats;
    }
  },
  methods: {
    sendMessage() {
      if (this.content !== "") {
        this.$store.dispatch("sendMessage", {
          user_id: this.ownID,
          content: this.content,
          chatID: this.chatID
        });
        this.content = "";
      }
    }
  }
};
</script>

<style scoped>
.chat-container {
  box-sizing: border-box;
  height: calc(100vh - 9.5rem);
  overflow-y: auto;
  padding: 10px;
  background-color: #f2f2f2;
}

.typer {
  justify-content: space-between;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  bottom: 0;
  height: 4.9rem;
  width: calc(100% - 20px);
  background-color: #fff;
  box-shadow: 0 -5px 10px -5px rgba(0, 0, 0, 0.2);
  position: absolute;
  padding: 15px;
}

.typer input[type="text"] {
  left: 2.5rem;
  background-color: transparent;
  border: none;
  outline: none;
  font-size: 1.25rem;
  height: 100%;
  width: 100%;
}
.container {
  height: 100%;
}
.inner {
  height: 100%;
}
</style>
