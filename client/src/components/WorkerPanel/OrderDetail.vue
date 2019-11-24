<template>
  <v-card>
    <v-card-title>注文詳細</v-card-title>
    <v-simple-table>
      <tbody>
        <tr>
          <td>注文者</td>
          <td>{{ order.user.display_name }}</td>
        </tr>
        <tr>
          <td>日程</td>
          <td>{{ order.date }}</td>
        </tr>
        <tr>
          <td>時間</td>
          <td>{{ order.time }}</td>
        </tr>
        <tr>
          <td>場所</td>
          <td>{{ order.location }}</td>
        </tr>
        <tr>
          <td>注文したもの</td>
          <td>
            <ul>
              <li
                v-for="(product, i) in order.products"
                :key="i"
              >{{ product.name }} × {{ product.stock }}</li>
            </ul>
          </td>
        </tr>
        <tr>
          <td>合計</td>
          <td>¥ {{ order.total_price }}</td>
        </tr>
      </tbody>
    </v-simple-table>
    <v-card-actions class="justify-center">
      <v-btn @click="onClickClose" color="grey lighten white--text">閉じる</v-btn>
      <v-btn
        color="light-blue  white--text"
        v-if="order.in_trade"
        @click="onClickFinishTrade"
      >対応済みにする</v-btn>
      <v-btn color="orange  white--text" v-else @click="onClickUnfinishTrade">対応済みを解除する</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>
export default {
  props: {
    orderDoc: {
      type: Object
    },
    onClickClose: {
      type: Function
    },
    onClickFinishTrade: {
      type: Function
    },
    onClickUnfinishTrade: {
      type: Function
    }
  },
  computed: {
    order: function() {
      return this.orderDoc.order;
    }
  }
};
</script>
