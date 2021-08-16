
if __name__ == "__main__":

    files = {"benign.csv": ["Nil", "Benign"], "defaced.csv": ["Medium", "Defaced"],
             "malware.csv": ["High", "Malware"], "phishing.csv": ["Medium", "Phishing"],
             "spam.csv": ["Low", "Spam"]}

    outfile = "init-db.js"
    template = '''
db.urlModel.insert({{
    \"url\": \"{0}\",
    \"risk\": \"{1}\",
    \"category\": \"{2}\"
}});
    '''

    with open(outfile, 'a') as outFile:
        for file in files:
            infile = "url_dataset/" + file
            risk = files[file][0]
            category = files[file][1]
            print(infile)
            with open(infile, 'r', encoding='utf-8') as inFile:
                for url in inFile:
                    sanitisedUrl = url.strip().replace("\"", "").replace("'", "").replace("\\", "")
                    outFile.write(template.format(sanitisedUrl, risk, category))
