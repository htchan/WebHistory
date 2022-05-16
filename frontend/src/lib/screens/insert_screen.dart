import 'package:flutter/material.dart';
import 'package:webhistory/models/popup.dart';
import 'package:webhistory/repostories/webHistoryRepostory.dart';

class InsertScreen extends StatefulWidget{
  final WebHistoryRepostory client;

  const InsertScreen({Key? key, required this.client}) : super(key: key);

  @override
  _InsertScreenState createState() => _InsertScreenState(this.client);
}

class _InsertScreenState extends State<InsertScreen> {
  WebHistoryRepostory client;
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  final TextEditingController urlTextController = TextEditingController();

  _InsertScreenState(this.client);

  String? validateUrl(String? url) {
    if (url == null || url.isEmpty) { return "Empty url"; }
    if (!url.startsWith("http")) { return "invalid url (not start with http)"; }
    return null;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History'),
      ),
      body: Form(
        key: _formKey,
        child: Column(
          children: [
            TextFormField(
              controller: urlTextController,
              decoration: const InputDecoration(hintText: "Url"),
              validator: validateUrl
            ),
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 16.0),
              child: ElevatedButton(
                onPressed: () {
                  if (_formKey.currentState!.validate()) {
                    client.createWeb(urlTextController.text)
                    .then( (popup) {
                      urlTextController.text = "";
                      popup?.show();
                    });
                  } else {
                    Popup("invald url").show();
                  }
                },
                child: const Text('Submit'),
              ),
            ),
          ],
        )
      )
    );
  }
}