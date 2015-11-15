package main

const TMPL = `
<!DOCTYPE html>
<html lang="en">
        <head>
                <meta charset="utf-8">
                <meta http-equiv="X-UA-Compatible" content="IE=edge">
                <meta name="viewport" content="width=device-width, initial-scale=1">
                <title>TPB</title>
                <!-- Latest compiled and minified CSS -->
                <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css" integrity="sha512-dTfge/zgoMYpP7QbHy4gWMEGsbsdZeCXz7irItjcC3sPUFtf0kuFbDz/ixG7ArTxmDjLXDmezHubeNikyKGVyQ==" crossorigin="anonymous">
        </head>
        <body>

                <div class="container">
                        <table class="table table-striped">
                                <thead>
                                        <tr>
                                                <th>#</th>
                                                <th>Cat</th>
                                                <th>Title</th>
                                                <th>Uploaded</th>
                                                <th>Size</th>
                                                <th>ULed</th>
                                                <th>Seeders</th>
                                                <th>Leechers</th>
                                                <th></th>
                                        </tr>
                                </thead>
                                <tbody>
                                        {{ range $i, $t := .}}
                                        <tr>
                                                <td>{{add $i 1}}</td>
                                                <td>{{$t.Category}}</td>
                                                <td><a href="{{.Magnet}}">{{$t.Title}}</a></td>
                                                <td>{{$t.Uploaded}}</td>
                                                <td>{{$t.Size}}</td>
                                                <td>{{$t.Uploader}}</td>
                                                <td>{{$t.Seeders}}</td>
                                                <td>{{$t.Leechers}}</td>
                                                <td>
                                                        <button class="btn btn-info btn-xs" data-clipboard-text="'{{$t.Magnet}}'">
                                                                copy
                                                        </button>
                                                </td>
                                        </tr>
                                        {{ end }}
                                </tbody>
                        </table>
                </div>

                <!-- 2. Include library -->
                <script src="https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/1.5.3/clipboard.min.js"></script>

                <!-- 3. Instantiate clipboard by passing a string selector -->
                <script>
                    var clipboard = new Clipboard('.btn');
                    clipboard.on('success', function(e) {
                        console.log(e);
                    });
                    clipboard.on('error', function(e) {
                        console.log(e);
                    });
                </script>
        </body>
</html>
`
