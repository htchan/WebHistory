// ignore: file_names
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:http/http.dart' as http;
import 'dart:html';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:url_launcher/url_launcher.dart';

class LoginScreen extends StatelessWidget {
  final Map queryParams;

  LoginScreen({Key? key, required this.queryParams}) : super(key: key) {
    String token = queryParams["token"] ?? "";
    if (token != "") {
      final Storage _localStorage = window.localStorage;
      _localStorage["web_history_token"] = token;
      redirect("/web-history/");
    } else {
      redirect(dotenv.env['USER_SERVICE_URL']!);
    }
  }

  void redirect(String url) {
    window.location.replace(url);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History Login'),
      ),
      body: Center(
        child: CircularProgressIndicator(),
      ),
    );
  }
}