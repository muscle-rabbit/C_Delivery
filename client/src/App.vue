<template>
  <v-app>
    <v-app-bar app color="grey darken-4" dark>
      <v-spacer></v-spacer>
    </v-app-bar>
    <router-view></router-view>
  </v-app>
</template>

<script>
export default {
  name: "App"
};
</script>
