using System;

namespace CSharpDemo
{
    class Program
    {
        static void Main(string[] args)
        {
            Console.WriteLine("Start.");

            SendTrans.SendNEOTransaction(8);
            //SendTrans.SendGASTransaction(11.13456789m);
            //SendTrans.SendNep5Transaction(21_00000000);

            Console.ReadKey();
        }       
        
    }
}
