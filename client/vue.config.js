module.exports = {
  transpileDependencies: ["vuetify"]
};

const webpack = require("webpack");
module.exports = {
  configureWebpack: {
    plugins: [
      new webpack.DefinePlugin({
        VUE_APP_API_ENDPOINT: JSON.stringify(
          process.env.VUE_APP_ENV === "production"
            ? "/"
            : "http://localhost:1964/"
        )
      })
    ]
  }
};
