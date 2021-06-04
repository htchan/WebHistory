import 'dart:convert';
// ignore: file_names
import 'package:flutter/material.dart';
// ignore: import_of_legacy_library_into_null_safe
import 'package:http/http.dart' as http;
import 'package:url_launcher/url_launcher.dart';

class MainPage extends StatefulWidget{
  final String url;

  const MainPage({Key? key, required this.url}) : super(key: key);

  @override
  _MainPageState createState() => _MainPageState(this.url);
}

class _MainPageState extends State<MainPage> {
  final String url;
  List<Widget> _web = [ const Center(child: Text("Loading")) ];
  // List<Widget> _buttons = _renderStageButton();
  final GlobalKey scaffoldKey = GlobalKey();

  _MainPageState(this.url) {
    final String apiUrl = '$url/list';
    http.get(Uri.parse(apiUrl))
    .then((response) {
      if (response.statusCode >= 200 && response.statusCode < 300) {
          Map<String, dynamic> body = Map.from(jsonDecode(response.body));
          List<Map<String, String>> websites = List.from(body['websites'])
          .map( (item) => Map<String, String>.from(item)).toList();
          setState(() {
            _web = websites.map(
              (website) => _renderWebsiteCard(websites.indexOf(website), website)
            ).toList();
          });
      } else {
        _web = [ const Center(child: Text("Failed to load data")) ];
      }
    });
  }

  Widget _renderWebsiteCard(int i, Map<String, String> website) {
    var accessTime = DateTime.parse(website['accessTime']??"20121225T0000");
    var updateTime = DateTime.parse(website['updateTime']??"20121225T0000");
    return GestureDetector(
      onTap: () async {
        final String apiUrl = '$url/refresh';
        http.post(
          Uri.parse(apiUrl),
          body: jsonEncode(<String, String>{
            'url': website['url']??"",
          }),
        );
        await canLaunch(website['url']??"")? await launch(website['url']??""):"";
      },
      child:ListTile(
        leading: (accessTime.millisecondsSinceEpoch > updateTime.millisecondsSinceEpoch) ? 
          const Icon(Icons.check_circle) : 
          const Icon(Icons.remove_circle),
        title: Text(website['title']??"Load title fail"),
        subtitle: Text((website['url']??"") + '\n' + updateTime.toLocal().toString()),
        trailing: IconButton(
          icon: const Icon(Icons.delete), 
          onPressed: () {
            final String apiUrl = '$url/delete';
            http.delete(
              Uri.parse(apiUrl),
              body: jsonEncode(<String, String>{
                'url': website['url'] == '' ? 'unknown' : (website['url']??"unknown")
              }),
            )
            .then((response) {
              if (response.statusCode >= 200 && response.statusCode < 300) {
                setState(() {
                  _web.removeAt(i);
                });
              }
            });
          },
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    // show the content
    return Scaffold(
      appBar: AppBar(
        title: const Text('Web History'),
        actions: [
          IconButton(
            onPressed: () {
              Navigator.pushNamed(
                scaffoldKey.currentContext!,
                '/add'
              );
            }, 
            icon: const Icon(Icons.add_circle),
          )
        ],
      ),
      key: scaffoldKey,
      body: Container(
        child: ListView.separated(
          separatorBuilder: (context, index) => const Divider(height: 10,),
          itemCount: _web.length,
          itemBuilder: (context, index) => _web[index],
        ),
        margin: const EdgeInsets.symmetric(horizontal: 5.0),
      ),
    );
  }
}