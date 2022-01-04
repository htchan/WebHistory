// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:http/http.dart' as http;
import 'dart:html';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:url_launcher/url_launcher.dart';

class LoginPage extends StatelessWidget {
  String token = "";

  LoginPage({Key? key}) : super(key: key) {
    String token = tokenCookie();
    print("cookie ${token}");
    if (token != "") {
      final Storage _localStorage = window.localStorage;
      _localStorage["web_history_token"] = token;
    } else {
      redirect(dotenv.env['USER_SERVICE_URL']!);
    }
  }
  
  String tokenCookie() {
    if (!(document.cookie ?? "").contains(';')) {
      return "";
    }
    Map<String, String>cookies = Map<String, String>.fromIterable(
      (document.cookie ?? "").split(";"),
      key: (item) => item.substring(0, item.indexOf('=')).trim(),
      value: (item) => item.substring(item.indexOf('=') + 1).trim
    );
    return cookies['token'] ?? "";
  }

  void redirect(String url) {
    window.location.href = url;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History Login'),
      ),
      body: Center(
        // child: CircularProgressIndicator(),
        child: Text(token),
      ),
    );
  }
}