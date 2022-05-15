// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:webhistory/Clients/webHistoryClient.dart';

class InsertPage extends StatefulWidget{
  WebHistoryClient client;

  InsertPage({Key? key, required this.client}) : super(key: key);

  @override
  _InsertPageState createState() => _InsertPageState(this.client);
}

class _InsertPageState extends State<InsertPage> {
  WebHistoryClient client;
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();

  _InsertPageState(this.client);

  String? validateUrl(String? url) {
    if (url == null || url.isEmpty) { return "Empty url"; }
    if (!url.startsWith("http")) { return "invalid url (not start with http)"; }
    return null;
  }

  void addUrl(TextEditingController text) {
    if (scaffoldKey.currentState!.validate()) {
      client.insert(text.text)
      .then( (popup) {
        text.text = "";
        resultToast(popup?.content?? "");
      });
    }
  }
  void resultToast(String msg) {
    Fluttertoast.showToast(
        msg: msg,
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

  @override
  Widget build(BuildContext context) {
    // show the content
    TextEditingController text = TextEditingController();
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History'),
      ),
      body: Form(
        key: scaffoldKey,
        child: Column(
          children: [
            TextFormField(
              controller: text,
              decoration: const InputDecoration(hintText: "Url"),
              validator: validateUrl
            ),
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16.0),
              child: ElevatedButton(
                onPressed: () => addUrl(text),
                child: const Text('Submit'),
              ),
            ),
          ],
        )
      )
    );
  }
}