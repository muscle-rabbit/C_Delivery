<template>
  <h1>welcome! redirecting...</h1>
</template>

<script>
export default {
  mounted() {
    const status = this.$route.query.status;
    const userID = this.$route.query.user;
    // eslint-disable-next-line no-console
    if (status === "worker") {
      this.$router.push(`/user/${userID}/worker_panel`);
    } else {
      const orderID = this.$route.query.order;
      const chatsID = this.$route.query.chats;
      this.$router.push(`/user/${userID}/order/${orderID}/chats/${chatsID}`);
    }
  }
};
</script>
