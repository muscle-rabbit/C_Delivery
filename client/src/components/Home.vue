<template>
  <h1>welcome!</h1>
</template>

<script>
export default {
  mounted() {
    const status = this.$route.query.status;
    const userID = this.$route.query.user;
    if (status === "worker") {
      // eslint-disable-next-line no-console
      console.log("this is worker");
      this.$router.push(`/user/${userID}/worker_panel`);
    } else {
      const orderID = this.$route.query.order;
      const chatsID = this.$route.query.chats;
      // eslint-disable-next-line no-console
      console.log("this is not worker");
      this.$router.push(`/user/${userID}/order/${orderID}/chats/${chatsID}`);
    }
  }
};
</script>
