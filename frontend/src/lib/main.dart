import 'package:flutter/material.dart';
import 'package:webhistory/Clients/webHistoryClient.dart';
import 'package:webhistory/Page/detailsPage.dart';
import 'dart:html';
import './Page/mainPage.dart';
import './Page/insertPage.dart';
import 'Page/loginPage.dart';
import 'package:flutter_web_plugins/flutter_web_plugins.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

class Strategy extends HashUrlStrategy {
  final PlatformLocation _platformLocation;

  Strategy([
    PlatformLocation _platformLocation = const BrowserPlatformLocation(),
  ]) : _platformLocation = _platformLocation,
      super(_platformLocation);

  @override
  String prepareExternalUrl(String internalUrl) {
    return internalUrl.isEmpty
      ? '${_platformLocation.pathname}${_platformLocation.search}'
      : '$internalUrl';
  }

  @override
  String getPath() {
    String path = _platformLocation.pathname + _platformLocation.search;
    if (!_platformLocation.hash.startsWith('#/')) {
      path += _platformLocation.hash;
    }
    return path;
  }
}

void main() async {
  await dotenv.load(fileName: ".env");
  setUrlStrategy(Strategy());
  runApp(MyApp());
}

String host = "192.168.128.146";
// String host = "localhost";
final Storage _localStorage = window.localStorage;

class MyApp extends StatelessWidget {
  // This widget is the root of your application.
  String url = 'http://${host}/api/web-history';
  WebHistoryClient client = WebHistoryClient("localhost", "");
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Web History',
      theme: ThemeData(
        textTheme: Theme.of(context).textTheme.apply(
          fontSizeFactor: 1.25,
        ),
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      initialRoute: '/',
      onGenerateRoute: (settings) {
        var uri = Uri.parse(settings.name??"");
        String authToken = _localStorage['web_history_token'] ?? "";
        if (authToken == "") {
          return MaterialPageRoute(builder: (context) => LoginPage(queryParams: uri.queryParameters), settings: settings);
        } else {
          client.authToken = authToken;
        }
        if (uri.pathSegments.indexOf('add') == 0) {
          return MaterialPageRoute(builder: (context) => InsertPage(client: client),
            settings: settings);
        } else if (uri.pathSegments.indexOf('details') == 0) {
          String groupName = uri.queryParameters["groupName"]??"";
          print("going to ${groupName}");
          return MaterialPageRoute(builder: (context) => DetailsPage(client: client, groupName: groupName),
            settings: settings);
        } else if (uri.path.startsWith('/web-history/user-service/login')) {
          return MaterialPageRoute(builder: (context) => LoginPage(queryParams: uri.queryParameters), settings: settings);
        } else {
          return MaterialPageRoute(builder: (context) => MainPage(client: client),
            settings: settings);
        }
      }
    );
  }
}

/*
http://host/                                            => main page
http://host/sites/<site>                                      => site page
http://host/search/<site>?title=<title>,writer=<writer> => search page
http://host/random/<site>                               => random page
http://host/books/<site>/<num>                                => book page
http://host/books/<site>/<num>/<version>                      => book page
*/
