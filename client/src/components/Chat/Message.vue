<template>
  <div id="message-content">
    <div
      class="message"
      v-for="(message, index) in chats.messages"
      :class="{ own: message.user_id == ownID }"
      :key="index"
    >
      <div class="inner" :class="message.user_id == ownID ? { own: true } : { other: true }">
        <div style="margin-top: 5px"></div>
        <div class="content" :class="message.user_id == ownID ? { own: true } : { other: true }">
          <div v-html="message.content"></div>
        </div>
        <div class="overline">{{ parseTime(message.created_at) }}</div>
      </div>
    </div>
  </div>
</template>

<script>
import moment from "moment";
export default {
  props: {
    ownID: {
      type: String,
      required: true
    },
    chats: {
      type: Object
    }
  },
  methods: {
    parseTime: function(time) {
      const parsedTime = new Date(Date.parse(time));
      return moment(parsedTime).format("HH:mm");
    }
  }
};
</script>

<style scoped>
.message {
  margin-bottom: 6px;
}
.message.own {
  text-align: right;
}

.message .content {
  background-color: lightgreen;
}
.displayName {
  font-size: 18px;
  font-weight: bold;
}

.content {
  padding: 8px;
  background-color: lightgreen;
  border-radius: 10px;
  display: inline-block;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.2), 0 1px 1px 0 rgba(0, 0, 0, 0.14),
    0 2px 1px -1px rgba(0, 0, 0, 0.12);
  max-width: 50%;
  word-wrap: break-word;
  background-color: lightskyblue;
}

.content.own {
  margin-left: 10px;
}
.content.other {
  margin-right: 10px;
}

.inner.own {
  display: flex;
  flex-direction: row-reverse;
  align-items: center;
}
.inner.other {
  display: flex;
  align-items: center;
}
</style>
