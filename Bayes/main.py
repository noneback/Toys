from utils import *

if __name__=='__main__':
    #获取词典并保存，一遍下次读取
    spam_voca=get_voca('./spam_voca.json',is_file_path=True)
    normal_voca=get_voca('./normal_voca.json',is_file_path=True)
    #保存方便下次读取
    save_voca2json(spam_voca,'./spam_voca.json')
    save_voca2json(normal_voca,'./normal_voca.json')
    #预处理测试数据
    s_x,s_y=get_files_labels('./data/test/spam/')
    n_x,n_y=get_files_labels('./data/test/normal/',is_spam=False)
    x,y=list(s_x+n_x),s_y+n_y
    #预测结果
    
    for i in range(20,40):
        print(str(i)+' : ',end='')
        predict_result(x,y,spam_voca,normal_voca,word_size=i)