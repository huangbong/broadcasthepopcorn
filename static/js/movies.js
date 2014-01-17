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
            url: "http://www.omdbapi.com/?s=" + t,
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
            url: "http://www.omdbapi.com/?i=" + imdbID,
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
        $('#download_container').show();        
        $('#download').html('Fetching download information from PTP... ' 
            + data + '...');
    }
});