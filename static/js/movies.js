/*
* movies.js
* Query OMDb API and PTP API to find/select movies and download correct torrent.
*/

$(document).ready(function() {
    var thread = null;

    $('#search').keyup(function() {
        clearTimeout(thread);
        var $this = $(this);
        clear();
        if (/\S/.test($this.val())) {
            thread = setTimeout(function() {
                getSearch($this.val())
            }, 600);
        }
    });

    function getSearch(t) {
        $.ajax({
            url: "http://www.omdbapi.com",
            data: {
                s: t
            },
            datatype: "json",
            success: function(data) {
                showResults($.parseJSON(data))
            }
        })
    }

    function clear() {
        $('#results_container').hide();
        $('#movie_container').hide();
        $('#download_container').hide();
        $('#download_button').hide();
        $('#results').empty();
        $('#search_error').empty();
    }

    function showResults(data) {
        var error = false;
        if (data['Response'] == "False") {
            error = true;
        }
        if (error == false) {
            $('#results_container').show();
            var search_html = '<ul>';
            var bad = 0;
            for (x in data['Search']) {
                if (data['Search'][x]['Type'] == 'movie') {
                    search_html += '<li id="r' + x + '"><a href="#movie">' 
                        + data['Search'][x]['Title'] + ' (' 
                        + data['Search'][x]['Year'] + ')' + '</a></li>';
                } else {
                    bad++;
                }
            }
            search_html += '</ul>';

            if (bad == data['Search'].length) {
                error = true;
            }
        }
        if (error == false) {
            $('#results').html(search_html);
            $('#results li a').click(function() {
                id = parseInt($(this).parent().attr('id').substring(1));
                imdbID = data['Search'][id]['imdbID'];
                getMovie(imdbID);

            });
        } else  {
            $('#results_container').show();
            if (typeof data['Error'] == 'undefined') {
                $('#search_error').html('Movie not found!');
            } else {
                $('#search_error').html(data['Error']);
            }           
        }
    }

    function getMovie(imdbID) {
        $.ajax({
            url: "http://www.omdbapi.com",
            data: {
                i: imdbID
            },
            datatype: "json",
            success: function(data) {
                showMovie($.parseJSON(data))
            }
        })
    }

    function showMovie(data) {
        $('#movie_container').show();
        $('.movie #title').html(data['Title']);
        if (data['Poster'] != 'N/A') {
            $('.movie #poster').attr('src', 'image?url=' + data['Poster']);
        } else {
            $('.movie #poster').attr('src', 'img/no_poster.gif');
        }
        $('.movie #year').html('(' + data['Year'] + ')');
        $('.movie #rating').html(data['Rated']);
        $('.movie #release').html(data['Released']);
        $('.movie #runtime').html(data['Runtime']);
        $('.movie #genre').html(data['Genre']);
        $('.movie #synopsis').html(data['Plot']);
        $('.movie #imdb').html('<a href="http://www.imdb.com/title/' 
            + data['imdbID'] + '" target="_blank">' + data['imdbID'] + '</a>');
        $('.movie #imdb_rating').html(data['imdbRating']);
        getDownload(data['imdbID']);
    }

    function getDownload(data) {
        $.ajax({
            url: "ptp_search",
            data: {
                imdbID: data
            },
            datatype: "json",
            success: function(data) {
                showDownload(data);
            }
        })        
    }

    function bytesToSize(bytes) {
        var sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB'];
        if (bytes == 0) return '0 Bytes';
        var i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
        return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + sizes[i];
    };

    function showDownload(data) {
        $('#download_container').show();
        $('#download_button').hide();
        if (data['Result'] == "OK") {
            var r = new Array(), j = -1;
            r[++j] = '<table>';
            r[++j] = '<tr><th>ID</th><th>Torrents</th><th>Size</th><th>Snatched</th> \
            <th>Seeders</th><th>Leechers</th></tr>';
            var quality;
            for (var key=0, size=data['Torrents'].length; key < size; key++) {
                if (data['Torrents'][key]['Quality'] != quality) {
                    r[++j] = '<tr class="border_bottom"><td colspan="6" class="quality">';
                    r[++j] = data['Torrents'][key]['Quality'];
                    r[++j] = '</td></tr>';
                    quality = data['Torrents'][key]['Quality'];
                }
                if(data['Torrents'][key]['Recommended']) {
                    r[++j] = '<tr class="recommended">';
                } else {
                    if (key % 2 == 1) {
                        r[++j] = '<tr class="grayrow">';
                    } else {
                        r[++j] = '<tr>';
                    }
                }
                r[++j] = '<td class="id">';
                r[++j] = data['Torrents'][key]['Id'];
                r[++j] = '</td><td class="torrents">';
                r[++j] = data['Torrents'][key]['Codec'];
                r[++j] = " / ";
                r[++j] = data['Torrents'][key]['Container'];
                r[++j] = " / ";
                r[++j] = data['Torrents'][key]['Source'];
                r[++j] = " / ";
                r[++j] = data['Torrents'][key]['Resolution'];
                if (data['Torrents'][key]['Scene']) {
                    r[++j] = " / "
                    r[++j] = "Scene";
                }
                if (data['Torrents'][key]['GoldenPopcorn']) {
                    r[++j] = ' <span class="goldenpopcorn">*</span>';
                }
                r[++j] = '</td><td class="size">';
                r[++j] = bytesToSize(data['Torrents'][key]['Size']);
                r[++j] = '</td><td class="snatched">';
                r[++j] = data['Torrents'][key]['Snatched'];
                r[++j] = '</td><td class="seeders">';
                r[++j] = data['Torrents'][key]['Seeders'];
                r[++j] = '</td><td class="leechers">';
                r[++j] = data['Torrents'][key]['Leechers'];
                r[++j] = '</td></tr>';
            }
            r[++j] = '</table>';
            $('#download').html(r.join(''));
            showDownloadButton(data['AuthKey'], data['PassKey']);

            $('tr').click(function() {
                $('tr.recommended').removeClass('recommended');
                $(this).addClass('recommended');
                showDownloadButton(data['AuthKey'], data['PassKey']);
            });

        } else {
            $('#download').html(data['Result']);
        }
    }

    function getTorrent(ptpID, authkey, passkey) {
        $.ajax({
            url: "ptp_get",
            data: {
                id: ptpID,
                authkey: authkey,
                passkey: passkey
            },
            datatype: "json",
            success: function(data) {
                showDownload(data);
            }
        })
    }

    function showDownloadButton(authkey, passkey) {
        $('#download_button button').attr("disabled", false);
        $('#download_button button').text('Download');
        $('#download_button').show();
        $('#download_button button').click(function() {
            $('#download_button button').attr("disabled", true);
            $('#download_button button').text('Adding download ' + $('.recommended .id').text() + '...');
            ptpID = $('.recommended .id').text();
            getTorrent(ptpID, authkey, passkey);
        });
    }

});