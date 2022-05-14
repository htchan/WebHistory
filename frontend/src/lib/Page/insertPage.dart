// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;

class InsertPage extends StatefulWidget{
  final String url;
  final String token;

  const InsertPage({Key? key, required this.url, required this.token}) : super(key: key);

  @override
  _InsertPageState createState() => _InsertPageState(this.url, this.token);
}

class _InsertPageState extends State<InsertPage> {
  final String url;
  final String token;
  List<Widget> _web = [ const Center(child: Text("Loading")) ];
  // List<Widget> _buttons = _renderStageButton();
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();

  _InsertPageState(this.url, this.token);

  String? validateUrl(String? url) {
    if (url == null || url.isEmpty) { return "Empty url"; }
    if (!url.startsWith("http")) { return "invalid url (not start with http)"; }
    return null;
  }

  void addUrl(TextEditingController text) {
    if (scaffoldKey.currentState!.validate()) {
      String apiUrl = '$url/websites/';
      // add loading animate
      http.post(
        Uri.parse(apiUrl),
        body: <String, String>{
          'url': text.text
        },
        headers: {"Authorization": token}
      )
      .then( (response) {
        // remove loading animate
        if (response.statusCode >= 200 && response.statusCode < 300) {
          text.text = "";
        }
        var data = response.body;
        resultToast(jsonDecode(data)["message"]);
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