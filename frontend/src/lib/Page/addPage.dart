// ignore: file_names
import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;

class AddPage extends StatefulWidget{
  final String url;

  const AddPage({Key? key, required this.url}) : super(key: key);

  @override
  _AddPageState createState() => _AddPageState(this.url);
}

class _AddPageState extends State<AddPage> {
  final String url;
  List<Widget> _web = [ const Center(child: Text("Loading")) ];
  // List<Widget> _buttons = _renderStageButton();
  final GlobalKey<FormState> scaffoldKey = GlobalKey<FormState>();

  _AddPageState(this.url);

  String? validateUrl(String? url) {
    if (url == null || url.isEmpty) { return "Empty url"; }
    if (!url.startsWith("http")) { return "invalid url (not start with http)"; }
    return null;
  }

  void addUrl(TextEditingController text) {
    if (scaffoldKey.currentState!.validate()) {
      String apiUrl = '$url/add';
      http.post(
        Uri.parse(apiUrl),
        body: <String, String>{
          'url': text.text
        }
      )
      .then( (response) {
        if (response.statusCode >= 200 && response.statusCode < 300) {
          text.text = "";
          successToast("url <${text.text}> add success");
        } else {
          successToast("fail");
        }
      });
    }
  }
  void successToast(String msg) {
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