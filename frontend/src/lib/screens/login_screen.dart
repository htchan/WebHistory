import 'package:flutter/material.dart';
import 'dart:html';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class LoginScreen extends StatelessWidget {
  final Map queryParams;
  final String FE_ROUTE_PREFIX = dotenv.get("WEB_WATCHER_FE_ROUTE_PREFIX", fallback: "/web-watcher");

  LoginScreen({Key? key, required this.queryParams}) : super(key: key) {
    String token = queryParams["token"] ?? "";
    if (token != "") {
      final Storage _localStorage = window.localStorage;
      _localStorage["web_history_token"] = token;
      redirect("${FE_ROUTE_PREFIX}/");
    } else {
      redirect(dotenv.get('USER_SERVICE_URL'));
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