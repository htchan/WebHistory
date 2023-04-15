import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';

class Popup {
  final String content;

  Popup(this.content);
  Popup.from(Map<String, String> map):
    this.content = map['message']??map['error']??'unknown message';

  void show() {
    Fluttertoast.showToast(
        msg: content,
        toastLength: Toast.LENGTH_LONG,
        gravity: ToastGravity.BOTTOM,
        timeInSecForIosWeb: 5,
        fontSize: 16.0,
        backgroundColor: Colors.grey.shade300,
        textColor: Colors.black,
        webBgColor: "#DDDDDD",
        webPosition: "center",
    );
  }
}