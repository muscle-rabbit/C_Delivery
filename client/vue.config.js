module.exports = {
  transpileDependencies: ["vuetify"],
  outputDir: "../server/dist",
  devServer:
    process.env.NODE_ENV === "development"
      ? {
          proxy: {
            "^/api": {
              target: "http://localhost:1964"
            }
          }
        }
      : null
};
