<template>
  <h1>welcome! redirecting...</h1>
</template>

<script>
export default {
  mounted() {
    const status = this.isUndefined(this.$route.query.status);
    const userID = this.isUndefined(this.$route.query.user);
    // eslint-disable-next-line no-console
    if (status === "worker") {
      this.$router.push(`/user/${userID}/worker_panel`);
    } else {
      const orderID = this.isUndefined(this.$route.query.order);
      const chatsID = this.isUndefined(this.$route.query.chats);
      this.$router.push(`/user/${userID}/order/${orderID}/chats/${chatsID}`);
    }
  },
  methods: {
    isUndefined(text) {
      if (text) {
        return text;
      }
      this.$router.push(`/NotFound`);
    }
  }
};
</script>
