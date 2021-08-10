import 'package:flutter/material.dart';
import 'package:webhistory/Page/detailsPage.dart';
import './Page/mainPage.dart';
import './Page/insertPage.dart';
import 'package:flutter_web_plugins/flutter_web_plugins.dart';

void main() {
  setUrlStrategy(PathUrlStrategy());
  runApp(MyApp());
}

// String host = "192.168.128.146";
String host = "192.168.128.146";

class MyApp extends StatelessWidget {
  // This widget is the root of your application.
  String url = 'http://${host}/api/web-history';
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
        print(uri.pathSegments);
        if (uri.pathSegments.indexOf('add') == 0) {
          return MaterialPageRoute(builder: (context) => InsertPage(url: url,),
            settings: settings);
        } else if (uri.pathSegments.indexOf('details') == 0) {
          String groupName = uri.queryParameters["groupName"]??"";
          print("going to ${groupName}");
          return MaterialPageRoute(builder: (context) => DetailsPage(url: url, groupName: groupName),
            settings: settings);
        } else {
          return MaterialPageRoute(builder: (context) => MainPage(url: url,),
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
