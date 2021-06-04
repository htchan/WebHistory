import 'package:flutter/material.dart';
import './Page/mainPage.dart';
import './Page/addPage.dart';
import 'package:flutter_web_plugins/flutter_web_plugins.dart';

void main() {
  setUrlStrategy(PathUrlStrategy());
  runApp(MyApp());
}


class MyApp extends StatelessWidget {
  // This widget is the root of your application.
  String url = 'http://localhost:9105/api/web-history';
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
        if (uri.pathSegments.indexOf('add') == 0) {
          return MaterialPageRoute(builder: (context) => AddPage(url: url,),
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
